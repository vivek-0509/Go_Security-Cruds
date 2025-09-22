package security

import (
	"awesomeProject/internal/handlers"
	"net/http"
)

func RegisterRoutes(taskHandler *handlers.TaskHandler) {
	// Public
	http.HandleFunc("/auth/google/callback", GoogleCallbackHandler)
	http.HandleFunc("/auth/google/login", GoogleLoginHandler)
}
