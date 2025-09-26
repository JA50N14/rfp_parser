package main

import (
	"errors"
	"log/slog"
	"os"
	"fmt"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	bearerTokenSmartsheet string
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
		os.Exit(1)
	}

	smartsheetRows, err := resultsToSmartsheetRows(allResults)

	fmt.Println(smartsheetRows)
	fmt.Println(allResults)

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

	extMap := map[string]string{
		".doc":  ".doc",
		".docx": ".docx",
		".xls":  ".xls",
		".xlsx": ".xlsx",
		".pdf":  ".pdf",
	}

	cfg := &apiConfig{
		bearerTokenSmartsheet: bearerTokenSmartsheet,
		rfpPackageRootDir:     RfpPackageRootDir,
		extMap:                extMap,
		logger:                logger,
	}

	return cfg, nil
}
