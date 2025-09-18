package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"regexp"
	"strings"
)

type KpiTracker struct {
	Name      string         `json:"name"`
	Category  string         `json:"category"`
	ColumnID  int            `json:"columnID"`
	Regexps   *regexp.Regexp `json:"-"`
	RegexStrs []string       `json:"regexps"` //temporary holder
	Found     bool           `json:"found"`
	Sentences []string       `json:"-"`
}

type SmartsheetRow struct {
	ToTop bool `json:"toTop"`
	Cells []struct {
		ColumnID int    `json:"columnID"`
		Value    string `json:"value"`
	} `json:"cells"`
}

const (
	kpiTrackerPath = "./kpiTracker.json"
)

func (cfg *apiConfig) traverseRfpPackages() error {
	rfpPackages, err := os.ReadDir(cfg.rfpPackageRootDir)
	if err != nil {
		cfg.logger.Error("Error reading RFP Packages in root directory", "Error", err)
		return err
	}

	smartSheetRows := []SmartsheetRow{}

	for _, rfpPackage := range rfpPackages {
		absPath := path.Join(cfg.rfpPackageRootDir, rfpPackage.Name())

		if !rfpPackage.IsDir() {
			cfg.logger.Info("File in the RFP Packages root directory.", "Filename", rfpPackage.Name())
			continue
		}

		rfpProcessedStatus, err := rfpProcessedCompleteCheck(absPath)
		if err != nil {
			cfg.logger.Error("Error checking if RFP Package has been parsed.", "Directory Name", rfpPackage.Name())
			return err
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
	return nil
}

func (cfg *apiConfig) traverseRfpPackage(rfpPackage string, smartSheetRows []SmartsheetRow) error {
	kpiTrackers, err := loadKpiTracker()
	if err != nil {
		cfg.logger.Error("Error loading kpiTracker", "Error", err)
		os.Exit(1)
	}

	kpiTrackers, err = compileRegexStrings(kpiTrackers)
	if err != nil {
		cfg.logger.Error("Error compiling regex strings to Regexp", "Error", err)
		os.Exit(1)
	}

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

func loadKpiTracker() ([]KpiTracker, error) {
	jsonFile, err := os.Open(kpiTrackerPath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	var kpiTrackers []KpiTracker

	err = decoder.Decode(&kpiTrackers)
	if err != nil {
		return nil, err
	}

	if len(kpiTrackers) == 0 {
		return nil, errors.New(".json file does not contain KPI parsing content.")
	}

	return kpiTrackers, nil
}

func compileRegexStrings(kpiTrackers []KpiTracker) ([]KpiTracker, error) {
	for i := range kpiTrackers {
		pattern := strings.Join(kpiTrackers[i].RegexStrs, "|")
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		kpiTrackers[i].Regexps = re
	}
	return kpiTrackers, nil
}
