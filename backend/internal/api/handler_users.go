package api

import (
	"RealTimeChat/backend/internal/database"
	"RealTimeChat/backend/internal/helpers"
	"RealTimeChat/backend/internal/type/model"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"net/http"
	"time"
)

// @Summary Signup a new user
// @Description Creates a new user account with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body model.Users true "User signup payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 406 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/signup [post]
func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var userBody model.Users
	// Get user input
	err := json.NewDecoder(r.Body).Decode(&userBody)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(userBody.Name) == 0 || len(userBody.Email) == 0 {
		helpers.WriteJSONError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	if len(userBody.Password) < 8 {
		helpers.WriteJSONError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, err := database.GetUser(bson.ObjectID{}, userBody.Email)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if len(user) > 0 {
		helpers.WriteJSONError(w, http.StatusConflict, "user already exists")
		return
	}

	userId := bson.NewObjectID()
	accessToken, refreshToken := helpers.GenerateToken(userBody.Email, userId)
	nUser := model.Users{
		Id:           userId,
		Name:         userBody.Name,
		Password:     *helpers.HashPassword(userBody.Password),
		Email:        userBody.Email,
		Token:        &accessToken,
		RefreshToken: &refreshToken,
	}

	err = database.PostUser(nUser)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Response
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "user created successfully",
		"userId":  userId,
	})
}

// @Summary Login user
// @Description Authenticates user and returns session cookies
// @Tags auth
// @Accept json
// @Produce json
// @Param user body model.Users true "User login payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userPayload := model.Users{}
	err := json.NewDecoder(r.Body).Decode(&userPayload)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if userPayload.Email == "" || userPayload.Password == "" {
		helpers.WriteJSONError(w, http.StatusBadRequest, "missing credentials")
		return
	}

	user, err := database.GetUser(bson.ObjectID{}, userPayload.Email)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if len(user) == 0 {
		helpers.WriteJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	passwordIsValid, err := helpers.VerifyPassword(user[0].Password, userPayload.Password)
	if err != nil || !passwordIsValid {
		helpers.WriteJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, refreshToken := helpers.GenerateToken(user[0].Email, user[0].Id)
	if err := helpers.UpdatedAllToken(token, refreshToken, user[0].Id); err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, "failed to update tokens")
		return
	}

	// Send it back
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,

		// production:
		Secure: false,

		Expires: time.Now().Add(7 * 24 * time.Hour),
	})

	// Response
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "login successful",
		//"token":   token,
	})
}

// @Summary Logout user.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/logout [get]
func Logout(w http.ResponseWriter, r *http.Request) {
	// Delete the JWT cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,

		// production:
		Secure:  false,
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})

	json.NewEncoder(w).Encode(map[string]any{
		"message": "logout successful",
	})
}
