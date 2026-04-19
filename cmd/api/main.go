package main

import (
	"RealTimeChat/config"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"RealTimeChat/internal/server"
)

type templateHandler struct {
	once     sync.Once
	filename string
	template *template.Template
}

// handling the template from our server
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.once.Do(func() {
		t.template = template.Must(template.ParseFiles(t.filename))
	})

	t.template.Execute(w, req)
}

// take answer from AI API
func main() {
	// Load credential from environment file
	cfg := config.Load()

	// Call AI Gemini
	server.Init()
	// Load Redis Address
	server.InitRedis(cfg.RedisAddr)

	// Make every randomly generated number unique
	rand.Seed(time.Now().UnixNano())

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("cmd/web/assets"))))
	http.Handle("/", &templateHandler{filename: "cmd/web/index.html"})
	http.Handle("/chat", &templateHandler{filename: "cmd/web/chat.html"})
	http.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")
		if len(roomName) == 0 {
			http.Error(w, "Room name required!", http.StatusBadRequest)
			return
		}

		userName := r.URL.Query().Get("name")
		if len(userName) == 0 {
			http.Error(w, "User name required!", http.StatusBadRequest)
			return
		}

		realRoom := server.GetRoom(roomName)
		realRoom.ServeHTTP(w, r)
	})

	fmt.Println(`
	 ______     ______         ______     ______   __
	/\  ___\   /\  __ \       /\  __ \   /\  == \ /\ \
	\ \ \__ \  \ \ \/\ \   -  \ \  __ \  \ \  _-/ \ \ \
	 \ \_____\  \ \_____\  -   \ \_\ \_\  \ \_\    \ \_\
	  \/_____/   \/_____/       \/_/\/_/   \/_/     \/_/ 
	   Starting web server on:`, cfg.AppAddr)
	if err := http.ListenAndServe(cfg.AppAddr, nil); err != nil {
		log.Fatal(err)
	}
}
