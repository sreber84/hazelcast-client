package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	hazelcast "github.com/hazelcast/hazelcast-go-client"
)

func main() {
	// Kubernetes Service and Port for the Hazelcast cluster
	hazelcastService := os.Getenv("HAZELCAST_SERVICE") // Example: "hazelcast-service.namespace-example.svc.cluster.local:5701"
	mapName := os.Getenv("HAZELCAST_MAP_NAME")         // Map name from environment variable

	if hazelcastService == "" {
		log.Fatal("HAZELCAST_SERVICE environment variable not set")
	}

	if mapName == "" {
		log.Fatal("HAZELCAST_MAP_NAME environment variable not set")
	}

	// Hazelcast configuration
	config := hazelcast.Config{}
	config.Cluster.Network.SetAddresses(hazelcastService)

	// Connect to Hazelcast
	client, err := hazelcast.StartNewClientWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to start Hazelcast client: %v", err)
	}
	defer client.Shutdown(context.Background())

	// Get a map from Hazelcast
	myMap, err := client.GetMap(context.Background(), mapName)
	if err != nil {
		log.Fatalf("Failed to get map: %v", err)
	}

	// Start data operations in parallel
	done := make(chan struct{})
	go writeDataInLoop(myMap, done)
	go readDataInLoop(myMap, done)

	// Wait indefinitely
	select {}
}

func writeDataInLoop(myMap *hazelcast.Map, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			// Write data
			key := "example-key"
			value := fmt.Sprintf("Hello, Hazelcast! Time: %s", time.Now().Format(time.RFC3339))
			err := myMap.Set(context.Background(), key, value)
			if err != nil {
				log.Printf("Failed to write data: %v", err)
			} else {
				log.Printf("Written data to Hazelcast: %s = %s", key, value)
			}

			// Wait a second before the next write
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
			// Read data
			key := "example-key"
			value, err := myMap.Get(context.Background(), key)
			if err != nil {
				log.Printf("Failed to read data: %v", err)
			} else {
				log.Printf("Read data from Hazelcast: %s = %v", key, value)
			}

			// Wait a second before the next read
			time.Sleep(1 * time.Second)
		}
	}
}
