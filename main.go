package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

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
	RfpPackageRootDir = `/home/jason_macfarlane/rfp_testtt`
)

func main() {
	flag.Parse() 

	args := flag.Args()
	if len(args) < 3 {
		fmt.Println("Usage: go run ./ <year> <business unit> <division>")
		os.Exit(1)
	}

	packagesYear := args[0]
	businessUnit := strings.ToUpper(args[1])
	division := strings.Join(args[2:], " ")

	if !strings.HasPrefix(packagesYear, "20") {
		fmt.Println("Type a valid <year>")
		os.Exit(1)
	}

	if !strings.Contains(businessUnit, "ABS") && !strings.Contains(businessUnit, "RHS") && !strings.Contains(businessUnit, "FM") {
		fmt.Println("Type a valid <business unit> - ABS, RHS, FM")
		os.Exit(1)
	}


	cfg, err := newApiConfig()
	if err != nil {
		os.Exit(1)
	}

	allResults, err := cfg.traverseRfpPackages(packagesYear, businessUnit, division)
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
		logger.Error("SMARTSHEET_TOKEN environment variable not set")
		return nil, errors.New("")
	}

	smartsheetUrl := os.Getenv("SMARTSHEET_URL")
	if smartsheetUrl == "" {
		logger.Error("SMARTSHEET_URL environment variable not set")
		return nil, errors.New("")
	}

	extMap := map[string]string{
		".docx": ".docx",
		".xlsx": ".xlsx",
		".pdf": ",pdf",
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
