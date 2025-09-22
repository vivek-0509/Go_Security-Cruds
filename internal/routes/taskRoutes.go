package routes

import (
	"awesomeProject/internal/handlers"
	"awesomeProject/internal/security"
	"net/http"
)

func RegisterRoutes(handler *handlers.TaskHandler) {
	http.HandleFunc("/healthz", handler.HealthHandler)
	//http.HandleFunc("/tasks", handler.TasksHandler)
	//http.HandleFunc("/tasks/", handler.TaskByIDHandler)
	// Secure
	http.Handle("/tasks", security.JwtFilter(http.HandlerFunc(handler.TasksHandler)))
	http.Handle("/tasks/", security.JwtFilter(http.HandlerFunc(handler.TaskByIDHandler)))
}
