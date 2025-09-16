package main

import (
	"os"
	"path"
)

type SmartsheetRow struct {
	ToTop bool `json:"toTop"`
	Cells []struct {
		ColumnID int    `json:"columnID"`
		Value    string `json:"value"`
	} `json:"cells"`
}

func (cfg *apiConfig) traverseRfpPackages() error {
	rfpPackages, err := os.ReadDir(cfg.rfpPackageRootDir)
	if err != nil {
		cfg.logger.Error("Error reading RFP Packages in root directory", "error", err)
		os.Exit(1)
	}

	smartSheetRows := []SmartsheetRow{}

	for _, rfpPackage := range rfpPackages {
		absPath := path.Join(cfg.rfpPackageRootDir, rfpPackage.Name())

		if !rfpPackage.IsDir() {
			cfg.logger.Info("File in the RFP Packages root directory.", "filename", rfpPackage.Name())
			continue
		}

		rfpProcessedStatus, err := rfpProcessedCompleteCheck(absPath)
		if err != nil {
			cfg.logger.Error("Error checking if RFP Package has been parsed.", "Directory Name", rfpPackage.Name())
			continue
		}

		if rfpProcessedStatus {
			cfg.logger.Info("RFP Package already processed.", "Directory Name", rfpPackage.Name())
			continue
		}

		err = cfg.traverseRfpPackage(absPath, smartSheetRows)
		if err != nil {
			cfg.logger.Error("Error parsing RFP Package", "Directory Name", rfpPackage.Name())
			continue
		}
	}
}

func (cfg *apiConfig) traverseRfpPackage(rfpPackage string, smartSheetRows []SmartsheetRow) error {

}

func rfpProcessedCompleteCheck(rfpRootDir string) (bool, error) {
	rfpPackage, err := os.ReadDir(rfpRootDir)
	if err != nil {
		return false, err
	}
	for _, item := range rfpPackage {
		if !item.IsDir() && item.Name() == "__processed.txt" {
			return true, nil
		}
	}
	return false, nil
}
