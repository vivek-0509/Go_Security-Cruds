package security

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = os.Getenv("JWT_SECRET")

func JwtFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("Invalid Authorization header")
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			//validate signing method
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || token.Valid {
			fmt.Println("Invalid Authorization header2")
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		//Extract Claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			//Put Claims into context
			ctx := context.WithValue(r.Context(), "user", claims["sub"])
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
