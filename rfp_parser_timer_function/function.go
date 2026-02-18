package timer_function

import (
	"context"
	"log/slog"
	"log"
	"os"
	"fmt"
	"time"

	"github.com/Azure/azure-functions-go/azfunc"
	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/target"
	"github.com/JA50N14/rfp_parser/walk"
	"github.com/joho/godotenv"
)


func TimerTrigger(ctx context.Context, timer azfunc.TimerInfo) error {
	log.Println("Timer triggered at:", time.Now())
	return RunParser(ctx)
}


func RunParser(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if os.Getenv("ENV") == "local" {
		if err := godotenv.Load(".env"); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	cfg, err := config.NewApiConfig(logger)
	if err != nil {
		return fmt.Errorf("failed to initialize API config: %w", err)
	}

	results, err := walk.WalkDocLibrary(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to walk document library: %w", err)
	}

	if len(results) == 0 {
		cfg.Logger.Info("there are currently no new RFP Packages to process")
		return nil
	}

	smartsheetRows := target.PrepareResultsForSmartsheetRows(results)

	err = target.PostToSmartsheets(smartsheetRows, ctx, cfg)
	if err != nil {
		return fmt.Errorf("post to smartsheets failed: %w", err)
	}

	cfg.Logger.Info("RFP Package(s) successfully parsed and posted to smartsheets")
	return nil
}
