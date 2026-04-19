package main

import (
	"RealTimeChat/backend/config"
	_ "RealTimeChat/backend/docs"
	"RealTimeChat/backend/internal/api"
	"RealTimeChat/backend/internal/database"
	server2 "RealTimeChat/backend/internal/server"
	"fmt"
	httpSwagger "github.com/swaggo/http-swagger"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var cfg *config.Config

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

// @title         RealTime Chat API
// @version		  1.0
// @description   This is a real-time chat service API.
// @contact.name  DoctorVeRossi
// @contact.url   https://github.com/vrstelios/RealTimeChat
// @BasePath      /
func main() {
	// Make every randomly generated number unique
	rand.Seed(time.Now().UnixNano())

	http.Handle("/swagger/", httpSwagger.WrapHandler)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/web/assets"))))
	http.Handle("/", &templateHandler{filename: "frontend/web/index.html"})
	http.Handle("/chat", &templateHandler{filename: "frontend/web/chat.html"})

	http.HandleFunc("/room", api.RoomHandler)

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

func init() {
	// Load credential from environment file
	cfg = config.Load()
	// Call AI Gemini
	server2.Init()
	// Load Redis Address
	server2.InitRedis(cfg.RedisAddr)
	// Load MongoDB Address
	database.InitDatabase(cfg.MongoURI)
}
