package api

import (
	"RealTimeChat/internal/server"
	"net/http"
)

// @Summary Join a chat room via WebSocket
// @Description Upgrades HTTP connection to WebSocket and joins a chat room
// @Tags chat
// @Param room  query string true  "Room name"
// @Param name  query string true  "Username"
// @Param useAI query bool   false "Enable AI responses (Gemini)"
// @Success 101 {string} string "Switching Protocols — WebSocket connection established"
// @Failure 400 {string} string "Bad Request — missing room or username"
// @Failure 500 {string} string "Internal Server Error — WebSocket upgrade failed"
// @Router /room [get]
func RoomHandler(w http.ResponseWriter, r *http.Request) {
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
}
