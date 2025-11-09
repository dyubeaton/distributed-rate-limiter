package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyubeaton/distributed-rate-limiter/internal/storage"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Logger Setup")

	//Start redis client
	redisClient, err := storage.NewRedisClient(storage.RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Timeout:  5 * time.Second,
	})

	if err != nil {
		log.Fatalf("Failed to connect to Redis %v", err)
	}

	defer redisClient.Close()

	log.Println("Connected to Redis")

	//create new mux to process requests and find the right handler
	mux := http.NewServeMux()

	/*
		The second argument cannot just be healthHandler because it doesn't match the signature of a HandlerFunction
		Remember, a HandlerFunction takes in a response writer and a request. Our HealthCheck function needs the redis client parameter as well.
		The solution is to create an anonymous function wrapper that takes in the same arguments as the Handler function and then internally calls
		our health check function, passing in the redis client we created. Very useful pattern.
	*/
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthHandler(w, r, redisClient)
	})

	//root endpoint for testing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"Rate Limiter API","version":"0.1.0"}`)
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	//call server in a goroutine, otherwise we're blocked and cant set up shutdown logic
	go func() {
		log.Printf("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)                      //channel of type os.Signal that has buffer size 1
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //will relay provided signals to channel (if we didn't specify the channel types it would send all I believe?)
	<-quit                                               //blocks the main goroutine by waiting for a signal from the channel

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //give outstanding requests 30 seconds before shutdown
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")

}

// checks if services and their dependencies are healthy
func healthHandler(w http.ResponseWriter, r *http.Request, redisClient *storage.RedisClient) {
	//we use r.Context here because it sets the context to that of the request, this way if the client shuts down, the context detects this, and frees up our resources
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

	defer cancel()

	redisHealthy := true

	if err := redisClient.Ping(ctx); err != nil {
		log.Printf("Redis health check failed: %v", err)
		redisHealthy = false
	}

	w.Header().Set("Content-Type", "application/json")
	if redisHealthy {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","redis":"connected"}`)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","redis":"disconnected"}`)
	}

}
