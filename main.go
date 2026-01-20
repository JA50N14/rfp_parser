package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JA50N14/rfp_parser/internal/auth"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	bearerTokenSmartsheet string
	smartsheetUrl         string
	rfpPackageRootDir     string
	accessToken string
	accessTokenExpiresAt time.Time
	extMap                map[string]string
	logger                *slog.Logger
	client *http.Client
}

const (
	RfpPackageRootDir = `/home/jason_macfarlane/rfp_doc_library`
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := newApiConfig(logger)
	if err != nil {
		logger.Error("Failed to initialize API config", "error", err)
		os.Exit(1)
	}
	os.Exit(0)

	allResults, err := cfg.traverseRfpPackages()
	if err != nil {
		logger.Error("Failed to traverse RFP Packages", "error", err)
		os.Exit(1)
	}

	smartsheetRows := resultsToSmartsheetRows(allResults)

	err = cfg.postRequestSmartsheets(smartsheetRows)
	if err != nil {
		logger.Error("Failed to post to Smartsheets", "error", err)
		os.Exit(1)
	}

	cfg.logger.Info("RFP Packages successfully parsed and posted to Smartsheets")
	os.Exit(0)
}

func newApiConfig(logger *slog.Logger) (*apiConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("failed to load .env file", "error", err)
		return nil, fmt.Errorf("loading .env file: %w", err)
	}

	bearerTokenSmartsheet := os.Getenv("SMARTSHEET_TOKEN")
	if bearerTokenSmartsheet == "" {
		logger.Error("SMARTSHEET_TOKEN not set in .env")
		return nil, fmt.Errorf("")
	}

	smartsheetUrl := os.Getenv("SMARTSHEET_URL")
	if smartsheetUrl == "" {
		logger.Error("SMARTSHEET_URL .env variable not set")
		return nil, errors.New("")
	}

	extMap := map[string]string{
		".docx": ".docx",
		".xlsx": ".xlsx",
	}

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	
	tokenResp, err := auth.GetGraphAccessToken(client)
	if err != nil {
		return nil, err
	}
	tokenExpiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	cfg := &apiConfig{
		bearerTokenSmartsheet: bearerTokenSmartsheet,
		smartsheetUrl:         smartsheetUrl,
		rfpPackageRootDir:     RfpPackageRootDir,
		extMap:                extMap,
		logger:                logger,
		client: client,
		accessToken: tokenResp.AccessToken,
		accessTokenExpiresAt: tokenExpiresAt,
	}
	return cfg, nil
}
