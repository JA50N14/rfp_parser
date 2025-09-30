package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	bearerTokenSmartsheet string
	smartsheetUrl         string
	clientSecretGraphApi  string
	clientIDGraphApi      string
	tenantIDGraphApi      string
	rfpPackageRootDir     string
	extMap                map[string]string
	logger                *slog.Logger
}

const (
	RfpPackageRootDir = `/home/jason_macfarlane/rfp_doc_library`
)

func main() {
	cfg, err := newApiConfig()
	if err != nil {
		os.Exit(1)
	}

	allResults, err := cfg.traverseRfpPackages()
	if err != nil {
		fmt.Printf("EXITING PROGRAM\n")
		os.Exit(1)
	}

	smartsheetRows := resultsToSmartsheetRows(allResults)
	err = cfg.postRequestSmartsheets(smartsheetRows)
	if err != nil {
		os.Exit(1)
	}

	cfg.logger.Info("RFP Packages Successfully Parsed and Posted to Smartsheets")
	os.Exit(0)
}

func newApiConfig() (*apiConfig, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error(".env file unreadable", "error", err)
		return nil, err
	}

	bearerTokenSmartsheet := os.Getenv("SMARTSHEET_TOKEN")
	if bearerTokenSmartsheet == "" {
		logger.Error("SMARTSHEET_TOKEN environment variable not set")
		return nil, errors.New("")
	}

	smartsheetUrl := os.Getenv("SMARTSHEET_URL")
	if smartsheetUrl == "" {
		logger.Error("SMARTSHEET_URL environment variable not set")
		return nil, errors.New("")
	}

	extMap := map[string]string{
		".doc":  ".doc",
		".docx": ".docx",
		".xls":  ".xls",
		".xlsx": ".xlsx",
		".pdf":  ".pdf",
	}

	cfg := &apiConfig{
		bearerTokenSmartsheet: bearerTokenSmartsheet,
		smartsheetUrl:         smartsheetUrl,
		rfpPackageRootDir:     RfpPackageRootDir,
		extMap:                extMap,
		logger:                logger,
	}

	return cfg, nil
}
