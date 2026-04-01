package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/walk"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Starting main")
	err := runParser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func runParser(ctx context.Context) error {
	fmt.Println("Starting parser job")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("Starting parser job")

	if os.Getenv("ENV") == "local" {
		if err := godotenv.Load(".env"); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	cfg, err := config.NewApiConfig(logger)
	if err != nil {
		return fmt.Errorf("failed to initialize API config: %w", err)
	}

	err = walk.WalkDocLibrary(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to walk document library: %w", err)
	}

	cfg.Logger.Info("RFP Package(s) successfully processed!")
	return nil
}
