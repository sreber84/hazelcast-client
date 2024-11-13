package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	hazelcast "github.com/hazelcast/hazelcast-go-client"
)

var client *hazelcast.Client
var myMap *hazelcast.Map

func main() {
	// Load configuration from environment variables
	hazelcastService := os.Getenv("HAZELCAST_SERVICE") // Example: "hazelcast.project-100.svc.cluster.local:5701"
	mapName := os.Getenv("HAZELCAST_MAP_NAME")         // Map name from environment variable

	if hazelcastService == "" || mapName == "" {
		log.Fatal("Environment variables HAZELCAST_SERVICE and HAZELCAST_MAP_NAME must be set")
	}

	// Hazelcast configuration
	config := hazelcast.Config{}
	config.Cluster.Network.SetAddresses(hazelcastService)

	// Connect to Hazelcast
	var err error
	client, err = hazelcast.StartNewClientWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to start Hazelcast client: %v", err)
	}
	defer client.Shutdown(context.Background())

	// Get a map from Hazelcast
	myMap, err = client.GetMap(context.Background(), mapName)
	if err != nil {
		log.Fatalf("Failed to get map: %v", err)
	}

	// Start background data operations
	done := make(chan struct{})
	go writeDataInLoop(myMap, done)
	go readDataInLoop(myMap, done)

	// Start the HTTP server for health check
	http.HandleFunc("/health", healthCheckHandler)
	log.Println("Starting HTTP server for health checks on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// healthCheckHandler checks Hazelcast connectivity and returns a health status response.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check Hazelcast connectivity by attempting a simple operation
	_, err := myMap.Get(ctx, "example-key")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	// If the operation succeeds, return healthy status
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func writeDataInLoop(myMap *hazelcast.Map, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			key := "example-key"
			value := fmt.Sprintf("Hello, Hazelcast! Time: %s", time.Now().Format(time.RFC3339))
			err := myMap.Set(context.Background(), key, value)
			if err != nil {
				log.Printf("Failed to write data: %v", err)
			} else {
				log.Printf("Written data to Hazelcast: %s = %s", key, value)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func readDataInLoop(myMap *hazelcast.Map, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			key := "example-key"
			value, err := myMap.Get(context.Background(), key)
			if err != nil {
				log.Printf("Failed to read data: %v", err)
			} else {
				log.Printf("Read data from Hazelcast: %s = %v", key, value)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
