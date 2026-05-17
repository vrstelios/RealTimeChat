package middleware

import (
	"RealTimeChat/backend/internal/helpers"
	"context"
	"net/http"
	"strings"
)

func Authenticate(provider TokenProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					token = parts[1]
				}
			}

			if token == "" {
				cookieToken, err := r.Cookie("Authorization")
				if err == nil && cookieToken.Value != "" {
					token = cookieToken.Value
				}
			}

			if token == "" {
				token = r.URL.Query().Get("token")
			}

			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := provider.Validate(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), helpers.CtxUserId, claims.UserId)
			ctx = context.WithValue(ctx, helpers.CtxEmail, claims.Email)

			// Continue request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// check cookie (ή Authorization header)
		cookie, err := r.Cookie("Authorization")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// validate token
		_, err = NewJWTTokenProvider().Validate(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
