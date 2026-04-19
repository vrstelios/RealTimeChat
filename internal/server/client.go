package server

import (
	"RealTimeChat/internal/database"
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
	streamId := fmt.Sprintf("gemini-%d", time.Now().UnixNano())

	var fullText strings.Builder

	// Take the message history from database
	history, _ := database.GetMessages(c.room.name)
	var allMsg []*genai.Content
	for _, h := range history {
		allMsg = append(allMsg, &genai.Content{
			Role:  h.Role,
			Parts: []*genai.Part{{Text: h.Message}},
		})
	}
	// Take the new message from user
	allMsg = append(allMsg, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: prompt}},
	})

	for result, err := range geminiClient.Models.GenerateContentStream(
		ctx,
		"gemini-3-flash-preview",
		allMsg,
		nil,
	) {
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				errMsg := Message{
					Name:    "Gemini",
					Message: "Rate limit.",
				}
				jsonErr, _ := json.Marshal(errMsg)
				c.receive <- jsonErr
			}
			log.Println("Stream error:", err)
			break
		}

		// Take the text from first candidate
		token := result.Candidates[0].Content.Parts[0].Text
		if token == "" {
			continue
		}

		fullText.WriteString(token)

		// Send any token
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

	// Done signal the client from asks
	doneMsg := Message{
		Name:      "Gemini",
		Message:   fullText.String(),
		Streaming: boolPtr(false),
		StreamId:  streamId,
	}
	jsonDone, _ := json.Marshal(doneMsg)
	select {
	case c.receive <- jsonDone:
	default:
		log.Println("Client buffer full on done:", c.name)
	}

	// Publish Redis the last message for all user in room
	// Without streamId in order to appear as normal message in the chat
	broadcastMsg := Message{
		Name:    "Gemini",
		Message: fullText.String(),
		Room:    c.name,
	}
	jsonBroadcast, _ := json.Marshal(broadcastMsg)
	redisCtx := context.Background()
	if err := c.room.rdb.Publish(redisCtx, "room:"+c.room.name, jsonBroadcast).Err(); err != nil {
		log.Println("Redis publish error:", err)
	}

	go database.SaveMessage(c.room.name, "Gemini", fullText.String(), "model")
}

func Init() {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	geminiClient = c
}
