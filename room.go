package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type room struct {

	// hold all current clients in the room
	clients map[*client]bool

	// join is a channel for all clients wishing to join this room
	join chan *client

	// leave is channel for all clients wishing to join this room
	leave chan *client

	// forward is a channel that holds incoming messages that should be forwarded to  other clients
	forward chan []byte
}

var rooms = make(map[string]*room)
var mu sync.Mutex

func getRoom(name string) *room {

	// prevent creating a room with the same name when multiple users do that at the same time
	mu.Lock()
	defer mu.Unlock()
	// if the room name already exists
	if r, ok := rooms[name]; ok {
		return r
	}
	// if not
	r := newRoom()
	rooms[name] = r

	go r.run()
	return r
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

// each room is a separate thread that should be run independently (but as long as the main server is running)
func (r *room) run() {
	for {
		select {
		// adding a user to a channel
		case cl := <-r.join:
			r.clients[cl] = true
		// removing a user from a channel
		case cl := <-r.leave:
			delete(r.clients, cl)
			close(cl.receive)
		// send a message to all clients in the room
		case msg := <-r.forward:
			for cl := range r.clients {
				cl.receive <- msg
			}
		}
	}
}

// upgrade a basic http connection to a websocket
const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upGrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	roomName := req.URL.Query().Get("room")
	if len(roomName) == 0 {
		http.Error(w, "Room name required", http.StatusBadRequest)
	}

	realRoom := getRoom(roomName)

	// Create socket
	socket, err := upGrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHTTP:", err)
		return
	}

	cl := &client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
		name:    fmt.Sprintf("User_%d", rand.Intn(10000)),
	}

	realRoom.join <- cl

	defer func() { r.leave <- cl }()

	go cl.write()
	cl.read()
}
