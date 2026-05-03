package server

import (
	"RealTimeChat/backend/internal/database"
	"RealTimeChat/backend/internal/mcp"
	"RealTimeChat/backend/internal/metrics"
	"RealTimeChat/backend/internal/rag"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/genai"
	"log"
	"strings"
	"time"
)

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Room    string `json:"room,omitempty"`

	// New field for AI usage
	Streaming *bool  `json:"streaming,omitempty"`
	StreamId  string `json:"streamId,omitempty"`
}

// Client is a single chatting user in a room
type Client struct {

	//a web socket for this user
	socket *websocket.Conn

	// receive is a channel to receive massages from other clients
	receive chan []byte

	room  *Room
	name  string
	useAI bool
}

var geminiClient *genai.Client

// Send messages function
func (c *Client) read() {
	// close the connection when we are done
	defer c.socket.Close()

	// as long as there is input, forward it
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		// added answer for individual
		outgoing := Message{
			Name:    c.name,
			Message: string(msg),
		}

		jsonMsg, err := json.Marshal(outgoing)
		if err != nil {
			fmt.Println("Encoding failed", err)
			continue
		}
		c.room.forward <- jsonMsg

		// added answer for AI agent
		if c.useAI {
			go c.streamGemini(strings.TrimPrefix(string(msg), "/ai "))
		}
	}
}

// Used to received messages
func (c *Client) write() {
	defer c.socket.Close()

	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

func boolPtr(b bool) *bool { return &b }

// Answer from Gemini api
func (c *Client) streamGemini(prompt string) {
	// Root span
	tracer := otel.Tracer("realtimechat")
	ctx, span := tracer.Start(context.Background(), "streamGemini")
	defer span.End()

	span.SetAttributes(
		attribute.String("room", c.room.name),
		attribute.String("user", c.name),
		attribute.String("prompt", prompt),
	)
	streamId := fmt.Sprintf("gemini-%d", time.Now().UnixNano())

	// Load chat history from MongoDB
	_, histSpan := tracer.Start(ctx, "load_history_mongodb")
	history, err := database.GetMessages(c.room.name)
	histSpan.End()
	if err != nil {
		log.Println("Failed to load history:", err)
	}

	// Build conversation history for Gemini
	var contents []*genai.Content
	for _, h := range history {
		contents = append(contents, &genai.Content{
			Role:  h.Role,
			Parts: []*genai.Part{{Text: h.Message}},
		})
	}

	// Add new user message
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: prompt}},
	})

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			mcp.SearchWebTool,
			mcp.SearchDocumentsTool,
		},
	}

	var fullText strings.Builder
	start := time.Now()

	// Tool calling loop
	for {
		// Child span — Gemini API call
		_, geminiSpan := tracer.Start(ctx, "gemini_api_call")

		resp, err := geminiClient.Models.GenerateContent(
			ctx,
			"gemini-3-flash-preview",
			contents,
			config,
		)
		if err != nil {
			geminiSpan.RecordError(err)
			geminiSpan.SetStatus(codes.Error, err.Error())
			geminiSpan.End()
			log.Println("Gemini error:", err)
			break
		}

		if len(resp.Candidates) == 0 {
			break
		}

		candidate := resp.Candidates[0]

		hasFunctionCall := false
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall == nil {
				continue
			}

			hasFunctionCall = true
			toolName := part.FunctionCall.Name
			args := part.FunctionCall.Args
			// Child span — Tool execution
			_, toolSpan := tracer.Start(ctx, "tool_call_"+toolName)
			toolSpan.SetAttributes(attribute.String("tool", toolName))

			var toolResult string
			switch toolName {
			case "search_web":
				query, _ := args["query"].(string)
				toolSpan.SetAttributes(attribute.String("query", query))
				toolResult, err = mcp.SearchWeb(query)
				if err != nil {
					toolSpan.RecordError(err)
					toolResult = "Web search failed: " + err.Error()
				}
			case "search_documents":
				query, _ := args["query"].(string)
				toolSpan.SetAttributes(attribute.String("query", query))

				queryEmbedding, err := rag.EmbedQuery(ctx, geminiClient, query)
				if err != nil {
					toolSpan.RecordError(err)
					toolResult = "Document search failed: " + err.Error()
					break
				}

				chunks, err := rag.SearchChunks(ctx, queryEmbedding, c.room.name, 5)
				if err != nil {
					toolSpan.RecordError(err)
					toolResult = "Document search failed: " + err.Error()
					break
				}

				toolSpan.SetAttributes(attribute.Int("chunks_found", len(chunks)))

				if len(chunks) == 0 {
					toolResult = "No relevant documents found."
				} else {
					toolResult = "Document context:\n\n" + strings.Join(chunks, "\n\n---\n\n")
				}
				log.Println("Document search found", len(chunks), "chunks")
			}

			toolSpan.End()

			contents = append(contents, candidate.Content)
			contents = append(contents, &genai.Content{
				Role: "user",
				Parts: []*genai.Part{
					{
						FunctionResponse: &genai.FunctionResponse{
							Name:     toolName,
							Response: map[string]any{"result": toolResult},
						},
					},
				},
			})
		}

		if !hasFunctionCall {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					fullText.WriteString(part.Text)
				}
			}
			break
		}
	}

	span.SetAttributes(
		attribute.Float64("gemini_latency_seconds", time.Since(start).Seconds()),
		attribute.Int("response_length", len(fullText.String())),
	)

	// Metrics
	metrics.GeminiLatency.Observe(time.Since(start).Seconds())
	finalText := fullText.String()
	if finalText != "" {
		metrics.AIRequestsTotal.WithLabelValues("success").Inc()
		metrics.MessagesTotal.WithLabelValues(c.room.name, "gemini").Inc()
	} else {
		metrics.AIRequestsTotal.WithLabelValues("error").Inc()
	}

	// Word by word streaming
	words := strings.Fields(finalText)
	for i, word := range words {
		token := word
		if i < len(words)-1 {
			token += " "
		}
		streamMsg := Message{
			Name:      "Gemini",
			Message:   token,
			Streaming: boolPtr(true),
			StreamId:  streamId,
		}
		jsonMsg, _ := json.Marshal(streamMsg)
		select {
		case c.receive <- jsonMsg:
		default:
			log.Println("Client buffer full:", c.name)
		}
	}

	// Done signal
	doneMsg := Message{
		Name:      "Gemini",
		Message:   finalText,
		Streaming: boolPtr(false),
		StreamId:  streamId,
	}
	jsonDone, _ := json.Marshal(doneMsg)
	select {
	case c.receive <- jsonDone:
	default:
	}

	// Save to MongoDB
	_, saveSpan := tracer.Start(ctx, "save_to_mongodb")
	go database.SaveMessage(c.room.name, c.name, prompt, "user")
	go database.SaveMessage(c.room.name, "Gemini", finalText, "model")
	saveSpan.End()

	// Broadcast to others via Redis
	_, redisSpan := tracer.Start(ctx, "redis_broadcast")
	broadcastMsg := Message{
		Name:    "Gemini",
		Message: finalText,
		Room:    c.name,
	}
	jsonBroadcast, _ := json.Marshal(broadcastMsg)
	if err := c.room.rdb.Publish(context.Background(), "room:"+c.room.name, jsonBroadcast).Err(); err != nil {
		redisSpan.RecordError(err)
		metrics.RedisPublishErrors.Inc()
		log.Println("Redis publish error:", err)
	}
	redisSpan.End()
}

func GetGeminiClient() *genai.Client {
	if geminiClient == nil {
		log.Fatal("Gemini client not initialized")
	}
	return geminiClient
}

func Init() {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	geminiClient = c
}
