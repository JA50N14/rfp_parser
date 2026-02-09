package main

import (
	"fmt"
	"log/slog"
	"os"
	"context"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/walk"
	"github.com/JA50N14/rfp_parser/target"
	"github.com/joho/godotenv"
)


func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("failed to load .env file: %w", err)
		os.Exit(1)
	}

	cfg, err := config.NewApiConfig(logger)
	if err != nil {
		logger.Error("failed to initialize API config", "error", err)
		os.Exit(1)
	}

///////////////////////
	fmt.Printf("ACCESS TOKEN: %s\n", cfg.AccessToken)
	os.Exit(0)
///////////////////////
	ctx := context.Background()
	results, err := walk.WalkDocLibrary(ctx, cfg)
	if err != nil {
		logger.Error("failed to walk document library", "error", err)
		os.Exit(1)
	}
	
	smartsheetRows := target.ResultsToSmartsheetRows(results)


// 	smartsheetRows := resultsToSmartsheetRows(results)

// 	err = cfg.postRequestSmartsheets(smartsheetRows)
// 	if err != nil {
// 		logger.Error("Failed to post to Smartsheets", "error", err)
// 		os.Exit(1)
// 	}

// 	cfg.logger.Info("RFP Packages successfully parsed and posted to Smartsheets")
// 	os.Exit(0)

}



