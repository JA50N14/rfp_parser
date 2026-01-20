package main

import (
	"errors"
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
	cfg, err := newApiConfig()
	if err != nil {
		os.Exit(1)
	}

	allResults, err := cfg.traverseRfpPackages()
	if err != nil {
		os.Exit(1)
	}

	smartsheetRows := resultsToSmartsheetRows(allResults)

	err = cfg.postRequestSmartsheets(smartsheetRows)
	if err != nil {
		os.Exit(1)
	}

	cfg.logger.Info("RFP Packages successfully parsed and posted to Smartsheets")
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
		logger.Error("SMARTSHEET_TOKEN .env variable not set")
		return nil, errors.New("")
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
