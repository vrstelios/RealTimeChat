package api

import (
	"RealTimeChat/backend/internal/helpers"
	"encoding/json"
	"net/http"
)

func MeHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(helpers.CtxUserId)
	email := r.Context().Value(helpers.CtxEmail)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"userId": userId,
		"email":  email,
	})
}
