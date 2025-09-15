package main

import (
	"context"
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

	loc, err := cli.SendLocation(ctx, gowa.SendLocationParams{
		Phone:     "558388572816@s.whatsapp.net",
		Latitude:  "-23.55052",
		Longitude: "-46.633308",
	})
	if err != nil {
		panic(err)
	}
	println("Location sent to:", loc)
}
