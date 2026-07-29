package main

import (
	"context"
	"log"
	"os"

	"github.com/itGeek-rus/smart-grid.git/internal/app"
)

func main() {
	application, err := app.NewIngestion()
	if err != nil {
		log.Printf("failed to init ingestion: %v", err)
		os.Exit(1)
	}
	if err := application.Run(context.Background()); err != nil {
		log.Printf("ingestion stopped with error: %v", err)
		os.Exit(1)
	}
}
