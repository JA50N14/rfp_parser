package config

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JA50N14/rfp_parser/internal/auth"

)

type ApiConfig struct {
	BearerTokenSmartsheet string
	SmartsheetUrl         string
	RFPPackageRootDir     string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	GraphSiteID string
	GraphLibraryName string
	GraphDriveID string
	ExtMap                map[string]string
	Logger                *slog.Logger
	Client                *http.Client
}

const (
	RfpPackageRootDir = `/home/jason_macfarlane/rfp_doc_library`
)


func NewApiConfig(logger *slog.Logger) (*ApiConfig, error) {


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

	cfg := &ApiConfig{
		BearerTokenSmartsheet: bearerTokenSmartsheet,
		SmartsheetUrl:         smartsheetUrl,
		RFPPackageRootDir:     RfpPackageRootDir,
		ExtMap:                extMap,
		Logger:                logger,
		Client:                client,
		AccessToken:           tokenResp.AccessToken,
		AccessTokenExpiresAt:  tokenExpiresAt,
		GraphSiteID: graphSiteID,
		GraphLibraryName: graphLibraryName,
		GraphDriveID: graphDriveID,
	}
	return cfg, nil
}

