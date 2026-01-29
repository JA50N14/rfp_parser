package main

import (
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
	accessToken           string
	accessTokenExpiresAt  time.Time
	graphSiteID string
	graphLibraryName string
	graphDriveID string
	extMap                map[string]string
	logger                *slog.Logger
	client                *http.Client
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
///////////////////////
	fmt.Printf("ACCESS TOKEN: %s\n", cfg.accessToken)
	os.Exit(0)
///////////////////////

	results, err := cfg.traverseRfpPackages()
	if err != nil {
		logger.Error("Failed to traverse RFP Packages", "error", err)
		os.Exit(1)
	}

	smartsheetRows := resultsToSmartsheetRows(results)

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
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	bearerTokenSmartsheet := os.Getenv("SMARTSHEET_TOKEN")
	if bearerTokenSmartsheet == "" {
		return nil, fmt.Errorf("SMARTSHEET_TOKEN not set in .env")
	}

	smartsheetUrl := os.Getenv("SMARTSHEET_URL")
	if smartsheetUrl == "" {
		return nil, fmt.Errorf("SMARTSHEET_URL .env variable not set")
	}
	
	graphSiteID := os.Getenv("GRAPH_SITE_ID")
	if graphSiteID == "" {
		return nil, fmt.Errorf("SHAREPOINT_SITE_ID .env variable not set")
	}

	graphLibraryName := os.Getenv("GRAPH_LIBRARY_NAME")
	if graphLibraryName == "" {
		return nil, fmt.Errorf("GRAPH_LIBRARY_NAME .env variable not set")
	}

	graphDriveID := os.Getenv("GRAPH_DRIVE_ID")
	if graphDriveID == "" {
		return nil, fmt.Errorf("GRAPH_DRIVE_ID .env variable not set")
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
		client:                client,
		accessToken:           tokenResp.AccessToken,
		accessTokenExpiresAt:  tokenExpiresAt,
		graphSiteID: graphSiteID,
		graphLibraryName: graphLibraryName,
		graphDriveID: graphDriveID,
	}
	return cfg, nil
}
