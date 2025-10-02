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

//PDF Parser:
//Tried a few parsers - rsc/pdf, ledongthuc - ledongthuc works better than rsc/pdf
//rsc/pdf is old and does not handle decompression well or grouped objects on pdfs
//There is a commerical pdf parser that is maintained, but it is subscription based

//Issue with using ledongthuc:
//Can iterate through each page of the pdf and decompress, but the pdf page can potential take up a lot of memory and os steps in to "signal: killed"
//Cannot check pdf file size before deciding to to parse PDF, because pdf file size is just the size of the pdf file on disk while compressed.
//Size of pdf does not indicated how memory is needed when parsing (i.e. pdf may contain high-res images, complex object streams for fonts or vector graphics)
//So a small pdf on disk can be very expensive in RAM

//Alternative Strategies Assessed:
//Graph API can convert several document types to pdf before sending over the wire, but not the other way around

//Convert a pdf to .docx in my program and then parse it once a .docx
//There are command line tools that can do this - can call commands from my progam, but this won't work in a serverless environment (LibreOffice)
//Pandoc for text-heavy PDFs
//Very slow if needing to process thousands of PDFs at scale
//Cannot use in a serverless environment (Azure Functions, Azure Logic Apps) because serverless often restricts disk access and this conversion is memory heavy, so may fail (risky & unreliable)

//Send pdfs over the wire to Copilot and get text back
//Can do this via goRoutines to speed up process
//Issue:
//-Currently no direct public API for Copilot ("Work" mode)
//-Copilot is tightly integrated with our Microsoft 365 environment (i.e. excel, outlook). No SDK or Rest API for programmatic interaction with Copilot "Work"
//-Large or complex PDFs (images, tables, scanned content) may not be parsed accurately by basic text extractors
//-Deal with Rate limit throttling

//Azure AI Document Intelligence (Form Recognizer)
//Enterprise-grade support for bulk PDF parsing

//Proposal Team converts PDFs to .docx
//Program grabs excel & word docs via Graph API
//Add a metadata column to their doc library housing RFPs - flag to set after finished processing that RFP Package

//Proposal Team Questions:
//-How does your team get ahold of RFP Packages? Email? SalesForce?
//-Where are your RFP Packages housed? Are they housed in multiple locations? Organized by BU?
//-Does your team house completed RFP Packages and blank RFP Packages you recieve from the client separately?
//-Does each folder represent a RFP Package in its entirity?
//-Are their multiple documents typically associated with an RFP Package?
//-What document types (xls, pdf, docx) are typically in a standard RFP Package?

//Based off responses:
//-Could your team convert PDFs associated with a RFP Package to .docx?
//-Could your team house blank RFP's and completed RFP's separately? (i.e. place a sub-directory in the RFP Package directory that is prefixed with a "_" or prefixed with anything)
//-Can we add a column to your document library? - Will use as a flag to programmatically mark parsed RFP Packages
//-

//How the program will work:
//-Create a SharePoint Site with a document library with a "Parsed" column
//-Give Proposals team access to create a directory for a RFP Package and drop RFP documents into it
//-They will convert the main RFP file into a .docx prior to uploading
//-Program to run on a set schedule
//-Hit Graph API endpoint - filter to only grab "Parsed" == ""
//-Cycle through the RFP Packages
//-Cycle through documents within each RFP Package
//-Put data for each RFP Package into a data structure
//Move RFP Package data structure into a struct that aligns with the Smartsheet API
//Encode to JSON and POST to Smartsheets
//If POST to Smartsheets 200, POST to Graph API to change "Parsed" to "true" for all RFP Packages that were processed


