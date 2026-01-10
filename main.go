package main

import (
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

type templateHandler struct {
	once     sync.Once
	filename string
	template *template.Template
}

// handling the template from our server

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.once.Do(func() {
		t.template = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	t.template.Execute(w, req)
}

// added name from any user
// take answer from AI API
// go run main.go client.go room.go
func main() {

	// make every randomly generated number unique
	rand.Seed(time.Now().UnixNano())

	var addr = flag.String("addr", ":8080", "Addr of the app")
	flag.Parse()
	r := newRoom()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/", &templateHandler{filename: "index.html"})
	http.Handle("/chat", &templateHandler{filename: "chat.html"})
	http.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		roomName := r.URL.Query().Get("room")
		if len(roomName) == 0 {
			http.Error(w, "Room name required!", http.StatusBadRequest)
			return
		}
		realRoom := getRoom(roomName)
		realRoom.ServeHTTP(w, r)
	})

	go r.run()

	// start the web server
	log.Println("Starting web server on:", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal()
	}
}
