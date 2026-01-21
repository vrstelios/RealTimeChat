package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {

	// hold all current clients in the room
	clients map[*Client]bool

	// join is a channel for all clients wishing to join this room
	join chan *Client

	// leave is channel for all clients wishing to join this room
	leave chan *Client

	// forward is a channel that holds incoming messages that should be forwarded to  other clients
	forward chan []byte
}

var rooms = make(map[string]*Room)
var mu sync.Mutex

func GetRoom(name string) *Room {

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

func newRoom() *Room {
	return &Room{
		forward: make(chan []byte),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]bool),
	}
}

// each room is a separate thread that should be run independently (but as long as the main server is running)
func (r *Room) run() {
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

func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	roomName := req.URL.Query().Get("room")
	if len(roomName) == 0 {
		http.Error(w, "Room name required", http.StatusBadRequest)
		return
	}

	username := req.URL.Query().Get("name")
	if len(roomName) == 0 {
		http.Error(w, "User name required", http.StatusBadRequest)
		return
	}

	useAI := req.URL.Query().Get("useAI")
	realRoom := GetRoom(roomName)

	// Create socket
	socket, err := upGrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("ServeHTTP:", err)
		return
	}

	cl := &Client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    realRoom,
		name:    username,
		useAI:   useAI,
	}

	realRoom.join <- cl
	defer func() { realRoom.leave <- cl }()

	go cl.write()
	cl.read()
}
