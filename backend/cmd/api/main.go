package main

import (
	"RealTimeChat/backend/config"
	_ "RealTimeChat/backend/docs"
	"RealTimeChat/backend/internal/api"
	"RealTimeChat/backend/internal/database"
	"RealTimeChat/backend/internal/helpers"
	"RealTimeChat/backend/internal/mcp"
	"RealTimeChat/backend/internal/metrics"
	"RealTimeChat/backend/internal/middleware"
	"RealTimeChat/backend/internal/rag"
	"RealTimeChat/backend/internal/server"
	"RealTimeChat/backend/internal/tracing"
	"context"
	"errors"
	"fmt"
	httpSwagger "github.com/swaggo/http-swagger"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	mainMux := http.NewServeMux()

	rand.Seed(time.Now().UnixNano())
	docHandler := api.NewDocumentHandler(server.GetGeminiClient())
	// Swagger documentation http://localhost:8080/swagger/index.html
	mainMux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Initialization Tracing
	shutdown := tracing.InitTracing("realtimechat", cfg.JaegerEndpoint)
	defer shutdown(context.Background())

	// Design Frontend
	mainMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("frontend/web/assets"))))
	mainMux.Handle("/login", &templateHandler{filename: "frontend/web/login.html"})
	mainMux.Handle("/signup", &templateHandler{filename: "frontend/web/signup.html"})
	authPages := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.ServeFile(w, r, "frontend/web/index.html")
		case "/chat":
			http.ServeFile(w, r, "frontend/web/chat.html")
		default:
			http.NotFound(w, r)
		}
	}))
	mainMux.Handle("/", authPages)
	mainMux.Handle("/chat", authPages)

	apiMux := http.NewServeMux()

	// public Login/Signup routes
	apiMux.HandleFunc("/auth/signup", api.Signup)
	apiMux.HandleFunc("/auth/login", api.Login)
	apiMux.HandleFunc("/auth/logout", api.Logout)

	// Middleware
	tokenProvider := middleware.NewJWTTokenProvider()

	// Protected routes
	apiMux.Handle("/room", middleware.Authenticate(tokenProvider)(http.HandlerFunc(api.RoomHandler)))
	apiMux.Handle("/documents/upload", middleware.Authenticate(tokenProvider)(http.HandlerFunc(docHandler.UploadDocument)))
	apiMux.Handle("/documents", middleware.Authenticate(tokenProvider)(http.HandlerFunc(docHandler.ListDocuments)))
	apiMux.Handle("/auth/me", middleware.Authenticate(tokenProvider)(http.HandlerFunc(api.MeHandler)))

	// mount api under /api
	mainMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	srv := &http.Server{
		Addr:    cfg.AppAddr,
		Handler: mainMux,
	}
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Starting web server!
	go func() {
		fmt.Println(`
     ______     ______         ______     ______   __
    /\  ___\   /\  __ \       /\  __ \   /\  == \ /\ \
    \ \ \__ \  \ \ \/\ \   -  \ \  __ \  \ \  _-/ \ \ \
     \ \_____\  \ \_____\  -   \ \_\ \_\  \ \_\    \ \_\
      \/_____/   \/_____/       \/_/\/_/   \/_/     \/_/`)

		log.Printf("Starting web server on %s...", cfg.AppAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server forced to shutdown unexpectedly: %v", err)
		}
	}()

	sig := <-shutdownChan
	log.Printf("Shutdown signal [%v] received. Initiating graceful shutdown...", sig)

	// --- GRACEFUL SHUTDOWN ---

	// Timeout 10 second in server in order completed HTTP requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Stopping HTTP server from accepting new connections...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown forced failure: %v", err)
	} else {
		log.Println("HTTP server stopped successfully.")
	}

	// Close WebSocket connections, The rooms and clean-up Redis
	log.Println("Cleaning up active chat rooms, clients and Redis state...")
	server.ShutdownRooms()

	log.Println("Server exited cleanly. Code deployed safely!")
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
	// Check Google search
	_, err := mcp.SearchWeb("golang websockets")
	if err != nil {
		log.Println("Search error:", err)
	}
	// load metrics
	metrics.Init()
	// Load JWT key
	helpers.InitJWT(cfg.JwtKey)
}
