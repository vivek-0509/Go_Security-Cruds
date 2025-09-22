package main

import (
	"awesomeProject/internal/handlers"
	"awesomeProject/internal/infrastructure"
	"awesomeProject/internal/routes"
	"awesomeProject/internal/security"
	"awesomeProject/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	mongoConn, err := infrastructure.NewMongoFromEnv()
	if err != nil {
		log.Fatal("failed to connect to MongoDb:", err)
	}
	defer mongoConn.Close(context.Background())

	repo := infrastructure.NewMongoTaskRepository(mongoConn)
	svc := service.NewTaskService(repo)
	handler := handlers.NewTaskHandler(svc)
	routes.RegisterRoutes(handler)
	security.RegisterRoutes(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Println("Shutdown error:", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("Server running on port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
	<-idleConnsClosed
	log.Println("Server stopped")
}
