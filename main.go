package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/walk"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	cfg, err := config.NewApiConfig(logger)
	if err != nil {
		logger.Error("failed to initialize API config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	err = walk.WalkDocLibrary(ctx, cfg)
	if err != nil {
		logger.Error("failed to walk document library: %w", err)
		os.Exit(1)
	}

	cfg.Logger.Info("RFP Package(s) successfully processed!")
	os.Exit(0)
}
