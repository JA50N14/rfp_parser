package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/target"
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

	////////////////////////////////////////////////////////////
	fmt.Printf("ACCESS TOKEN: %s\n", cfg.AccessToken)
	os.Exit(0)
	//////////////////////////////////////////////////////////////

	ctx := context.Background()
	results, err := walk.WalkDocLibrary(ctx, cfg)
	if err != nil {
		cfg.Logger.Error("failed to walk document library", "error", err)
		os.Exit(1)
	}

	smartsheetRows := target.PrepareResultsForSmartsheetRows(results)

	err = target.PostToSmartsheets(smartsheetRows, ctx, cfg)
	if err != nil {
		cfg.Logger.Error("post to smartsheets failed", "error", err)
		os.Exit(1)
	}

	cfg.Logger.Info("RFP Packages successfully parsed and posted to smartsheets")
	os.Exit(0)
}
