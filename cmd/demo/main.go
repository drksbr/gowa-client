package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/drksbr/gowa-client/pkg/gowa"
)

func main() {
	ctx := context.Background()
	baseURL := os.Getenv("GOWA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	cli, err := gowa.New(gowa.Config{
		BaseURL:  baseURL,
		Username: os.Getenv("GOWA_USER"),
		Password: os.Getenv("GOWA_PASS"),
		Timeout:  20 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("[INFO] Testando login QR (GET /app/login)")
	login, err := cli.Login(ctx)
	if err != nil {
		fmt.Printf("[ERROR] Login: %v\n", err)
	} else {
		b, _ := json.MarshalIndent(login, "", "  ")
		fmt.Printf("[RESPONSE] Login: %s\n", b)
	}
}
