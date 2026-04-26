package main

import (
	"RealTimeChat/backend/config"
	_ "RealTimeChat/backend/docs"
	"RealTimeChat/backend/internal/api"
	"RealTimeChat/backend/internal/database"
	"RealTimeChat/backend/internal/mcp"
	"RealTimeChat/backend/internal/rag"
	server "RealTimeChat/backend/internal/server"
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
// @BasePath      /api
func main() {
	rand.Seed(time.Now().UnixNano())
	docHandler := api.NewDocumentHandler(server.GetGeminiClient())
	// Swagger documentation http://localhost:8080/swagger/index.html
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// Design Frontend
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/web/assets"))))
	http.Handle("/", &templateHandler{filename: "frontend/web/index.html"})
	http.Handle("/chat", &templateHandler{filename: "frontend/web/chat.html"})

	// Routes Backend
	http.HandleFunc("/room", api.RoomHandler)
	http.HandleFunc("/api/documents/upload", docHandler.UploadDocument)
	http.HandleFunc("/api/documents/", docHandler.ListDocuments)

	// Starting web server!
	fmt.Println(`
	 ______     ______         ______     ______   __
	/\  ___\   /\  __ \       /\  __ \   /\  == \ /\ \
	\ \ \__ \  \ \ \/\ \   -  \ \  __ \  \ \  _-/ \ \ \
	 \ \_____\  \ \_____\  -   \ \_\ \_\  \ \_\    \ \_\
	  \/_____/   \/_____/       \/_/\/_/   \/_/     \/_/`)
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
	// Load qdrant
	if err := rag.InitQdrant(cfg.QdrantHost, cfg.QdrantPort); err != nil {
		log.Fatal("Qdrant init failed:", err)
	}
	// Test Google search
	result, err := mcp.SearchWeb("golang websockets")
	if err != nil {
		log.Println("Search error:", err)
	} else {
		log.Println("Search result:", result)
	}
}
