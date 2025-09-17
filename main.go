package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//Initialize MongoDb
	mongoConn, err := NewMongoFromEnv()
	if err != nil {
		log.Fatal("failed to connect to MongoDb: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = mongoConn.Close(ctx)
	}()

	srv := NewServer(mongoConn)
	srv.Routes()

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{Addr: ":" + addr}

	//Graceful shutdown setup
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down... signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Println("failed to shutdown gracefully:", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("Starting server on", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
	<-idleConnsClosed
	log.Println("Server shut down")
}
