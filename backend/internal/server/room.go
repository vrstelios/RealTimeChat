package server

import (
	"RealTimeChat/backend/internal/database"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	name string
	// hold all current clients in the room
	clients map[*Client]bool

	// join is a channel for all clients wishing to join this room
	join chan *Client

	// leave is channel for all clients wishing to join this room
	leave chan *Client

	// forward is a channel that holds incoming messages that should be forwarded to  other clients
	forward chan []byte

	rdb *redis.Client
}

var (
	rooms = make(map[string]*Room)
	mu    sync.Mutex
	rdb   *redis.Client
)

func InitRedis(addr string) {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}
	log.Println("Redis connected:", addr)
}

func GetRoom(name string) *Room {

	// prevent creating a room with the same name when multiple users do that at the same time
	mu.Lock()
	defer mu.Unlock()
	// if the room name already exists
	if r, ok := rooms[name]; ok {
		return r
	}
	// if not
	r := newRoom(name)
	rooms[name] = r

	// Save room in Redis
	ctx := context.Background()
	rdb.HSet(ctx, "rooms", name, "active")

	go r.run()
	go r.subscribeRedis()
	return r
}

func newRoom(name string) *Room {
	return &Room{
		name:    name,
		forward: make(chan []byte),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]bool),
		rdb:     rdb,
	}
}

// Each room is a separate thread that should be run independently (but as long as the main server is running)
func (r *Room) run() {
	for {
		select {
		// adding a user to a channel
		case cl := <-r.join:
			r.clients[cl] = true
			// Save username from Redis
			ctx := context.Background()
			r.rdb.HSet(ctx, "room:"+r.name+":users", cl.name, "online")
		// removing a user from a channel
		case cl := <-r.leave:
			delete(r.clients, cl)
			close(cl.receive)
			// Remove username from Redis
			ctx := context.Background()
			r.rdb.HDel(ctx, "room:"+r.name+":users", cl.name)
		// send a message to all clients in the room
		case msg := <-r.forward:
			ctx := context.Background()
			if err := r.rdb.Publish(ctx, "room:"+r.name, msg).Err(); err != nil {
				log.Println("Redis publish error:", err)
			}
		}
	}
}

func (r *Room) subscribeRedis() {
	ctx := context.Background()
	sub := r.rdb.Subscribe(ctx, "room:"+r.name)
	defer sub.Close()

	ch := sub.Channel()
	for msg := range ch {
		var m Message
		// If Gemini broadcast send in all, however send missing message to all clients in the room
		if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		if m.Name != "Gemini" {
			// Just arrived message from Redis, save it to MongoDB
			go database.SaveMessage(r.name, m.Name, m.Message, "user")
		}

		// Broadcast the message to all clients in the room
		for cl := range r.clients {
			if m.Name == "Gemini" && m.Room == cl.name {
				continue
			}
			select {
			case cl.receive <- []byte(msg.Payload):
			default:
				// If the client's receive channel is full, skip sending the message
				log.Printf("Client %s receive channel full, skipping message\n", cl.name)
			}
		}
	}
}

// upgrade a basic http connection to a websocket
const (
	socketBufferSize  = 1024
	messageBufferSize = 512
)

var upGrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize, CheckOrigin: func(r *http.Request) bool { return true }}

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

	useAI := req.URL.Query().Get("useAI") == "true"
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

	history, err := database.GetMessages(roomName)
	if err != nil {
		log.Println("Failed to load history:", err)
	} else {
		for _, msg := range history {
			histMsg := Message{
				Name:    msg.Name,
				Message: msg.Message,
			}
			jsonMsg, _ := json.Marshal(histMsg)
			// Update the socket with old message before cl.write start yet
			if err = socket.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
				log.Println("Failed to send history:", err)
				break
			}
		}
	}

	go cl.write()
	cl.read()
}
