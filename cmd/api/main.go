package main

import (
	"RealTimeChat/config"
	_ "RealTimeChat/docs"
	"RealTimeChat/internal/api"
	"RealTimeChat/internal/database"
	"fmt"
	httpSwagger "github.com/swaggo/http-swagger"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"RealTimeChat/internal/server"
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

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("cmd/web/assets"))))
	http.Handle("/", &templateHandler{filename: "cmd/web/index.html"})
	http.Handle("/chat", &templateHandler{filename: "cmd/web/chat.html"})
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
	server.Init()
	// Load Redis Address
	server.InitRedis(cfg.RedisAddr)
	// Load MongoDB Address
	database.InitDatabase(cfg.MongoURI)
}
