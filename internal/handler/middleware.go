package handler

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. получи заголовок Authorization
		authHeader := r.Header.Get("Authorization")

		// 2. проверь что он не пустой и начинается с "Bearer "
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return

		}

		// 3. вырежи токен из строки "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString,
			func(token *jwt.Token) (any, error) {
				return []byte("secret_key"), nil
			})
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := int64(claims["user_id"].(float64))

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
