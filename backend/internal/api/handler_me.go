package api

import (
	"RealTimeChat/backend/internal/helpers"
	"encoding/json"
	"net/http"
)

// @Summary Get current user
// @Description Returns the authenticated user's info
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/me [get]
func MeHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(helpers.CtxUserId)
	email := r.Context().Value(helpers.CtxEmail)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"userId": userId,
		"email":  email,
	})
}
