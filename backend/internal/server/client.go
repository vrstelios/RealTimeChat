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
	ctx := context.Background()
	start := time.Now()
	streamId := fmt.Sprintf("gemini-%d", time.Now().UnixNano())

	// Load chat history from MongoDB
	history, err := database.GetMessages(c.room.name)
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

	// Tool calling loop
	for {
		resp, err := geminiClient.Models.GenerateContent(
			ctx,
			"gemini-3-flash-preview",
			contents,
			config,
		)
		if err != nil {
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
			log.Printf("Gemini calls tool: %s with args: %v\n", toolName, args)

			var toolResult string
			switch toolName {
			case "search_web":
				query, _ := args["query"].(string)
				toolResult, err = mcp.SearchWeb(query)
				if err != nil {
					toolResult = "Web search failed: " + err.Error()
				}
				log.Println("Web search result:", toolResult)

			case "search_documents":
				query, _ := args["query"].(string)

				queryEmbedding, err := rag.EmbedQuery(ctx, geminiClient, query)
				if err != nil {
					toolResult = "Document search failed: " + err.Error()
					break
				}

				chunks, err := rag.SearchChunks(ctx, queryEmbedding, c.room.name, 5)
				if err != nil {
					toolResult = "Document search failed: " + err.Error()
					break
				}

				if len(chunks) == 0 {
					toolResult = "No relevant documents found for this room."
				} else {
					toolResult = "Document context:\n\n" + strings.Join(chunks, "\n\n---\n\n")
				}
				log.Println("Document search found", len(chunks), "chunks")
			}

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

	metrics.GeminiLatency.Observe(time.Since(start).Seconds())

	// Word by word streaming
	finalText := fullText.String()
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

	if finalText != "" {
		metrics.AIRequestsTotal.WithLabelValues("success").Inc()
		metrics.MessagesTotal.WithLabelValues(c.room.name, "gemini").Inc()
	} else {
		metrics.AIRequestsTotal.WithLabelValues("error").Inc()
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
		log.Println("Client buffer full on done:", c.name)
	}

	// Save to MongoDB
	go database.SaveMessage(c.room.name, c.name, prompt, "user")
	go database.SaveMessage(c.room.name, "Gemini", finalText, "model")

	// Broadcast to others via Redis
	broadcastMsg := Message{
		Name:    "Gemini",
		Message: finalText,
		Room:    c.name,
	}
	jsonBroadcast, _ := json.Marshal(broadcastMsg)
	if err := c.room.rdb.Publish(ctx, "room:"+c.room.name, jsonBroadcast).Err(); err != nil {
		metrics.RedisPublishErrors.Inc()
		log.Println("Redis publish error:", err)
	}
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
