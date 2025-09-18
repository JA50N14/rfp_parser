package main

import (
	"log/slog"
	"os"
	"errors"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	bearerTokenSmartsheet string
	clientSecretGraphApi string
	clientIDGraphApi string
	tenantIDGraphApi string
	rfpPackageRootDir string
	fileExtensions []string
	logger *slog.Logger
}

const (
	RfpPackageRootDir = `C:\Users\Macfa\RFP_Packages`
	PdfExt = ".pdf"
	DocExt = ".doc"
	XlsExt = ".xls"
)

func main() {
	cfg, err := newApiConfig()
	if err != nil {
		os.Exit(1)
	}

	err = cfg.traverseRfpPackages()
	if err != nil {
		os.Exit(1)
	}



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

	cfg := &apiConfig {
		bearerTokenSmartsheet: bearerTokenSmartsheet,
		rfpPackageRootDir: RfpPackageRootDir,
		fileExtensions: []string{PdfExt, DocExt, XlsExt},
		logger: logger,
	}

	return cfg, nil 
}