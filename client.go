package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Room    string `json:"room,omitempty"`
}

// client is a single chatting user in a room
type client struct {

	//a web socket for this user
	socket *websocket.Conn

	// receive is a channel to receive massages from other clients
	receive chan []byte

	room *room
	name string
}

// send messages function
func (c *client) read() {
	// close the connection when we are done
	defer c.socket.Close()

	// as long as there is a input, forward it
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

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
	}
}

// used to received messages
func (c *client) write() {
	defer c.socket.Close()

	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
