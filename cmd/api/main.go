package main

import (
	"context"
	"log"
	"os"

	"github.com/itGeek-rus/smart-grid.git/internal/app"
)

func main() {
	application, err := app.NewAPI()
	if err != nil {
		log.Printf("failed to init api: %v", err)
		os.Exit(1)
	}
	if err := application.Run(context.Background()); err != nil {
		log.Printf("failed to run: %v", err)
		os.Exit(1)
	}
}
