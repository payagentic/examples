package main

import (
	"context"
	"fmt"
	payagentic "github.com/payagentic/payagentic-go"
	"log"
	"os"
)

func main() {
	baseURL := os.Getenv("PAYAGENTIC_BASE_URL")
	if baseURL == "" {
		log.Fatal("Set PAYAGENTIC_BASE_URL")
	}
	client, err := payagentic.NewClient(payagentic.WithBaseURL(baseURL))
	if err != nil {
		log.Fatal("Check SDK configuration")
	}
	defer client.Close()
	result, err := client.OpenAPI.ListWalletsWithResponse(context.Background(), nil)
	if err != nil {
		log.Fatal("Gateway request failed")
	}
	fmt.Println("Gateway HTTP status:", result.StatusCode())
}
