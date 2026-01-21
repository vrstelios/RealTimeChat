package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
	"log"
	"strings"
)

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Room    string `json:"room,omitempty"`
}

// client is a single chatting user in a room
type Client struct {

	//a web socket for this user
	socket *websocket.Conn

	// receive is a channel to receive massages from other clients
	receive chan []byte

	room  *Room
	name  string
	useAI string
}

var geminiClient *genai.Client

// send messages function
func (c *Client) read() {
	// close the connection when we are done
	defer c.socket.Close()

	// as long as there is a input, forward it
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
		if c.useAI == "true" {
			answer, err := callGemini(strings.TrimPrefix(string(msg), "/ai "))
			if err != nil {
				answer = "AI error"
			}

			outgoing2 := Message{
				Name:    "Gemini",
				Message: answer,
			}

			jsonMsg2, err := json.Marshal(outgoing2)
			if err != nil {
				fmt.Println("Encoding failed", err)
				continue
			}

			c.room.forward <- jsonMsg2
		}
	}
}

// used to received messages
func (c *Client) write() {
	defer c.socket.Close()

	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

// answer from Gemini api
func callGemini(msg string) (string, error) {
	ctx := context.Background()

	// The client gets the API key from the environment variable `GEMINI_API_KEY`.
	result, err := geminiClient.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text(msg),
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}

func Init() {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	geminiClient = c
}
