package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"regexp"
)

type KpiTrackerDef struct {
	Name      string         `json:"name"`
	Category  string         `json:"category"`
	ColumnID  int            `json:"columnID"`
	Regexps   []*regexp.Regexp `json:"-"`
	RegexStrs []string       `json:"regexps"` //temporary holder
}

type KpiResult struct {
	*KpiTrackerDef
	Found     bool
	Sentence  string
}

type SmartsheetRow struct {
	ToTop bool `json:"toTop"`
	Cells []struct {
		ColumnID int    `json:"columnID"`
		Value    string `json:"value"`
	} `json:"cells"`
}

const (
	kpiTrackerDefPath = "./kpiTracker.json"
)

func (cfg *apiConfig) traverseRfpPackages() ([]SmartsheetRow, error) {
	rfpPackages, err := os.ReadDir(cfg.rfpPackageRootDir)
	if err != nil {
		cfg.logger.Error("Error reading RFP Packages in root directory", "Error", err)
		return nil, err
	}

	kpiTrackerDefs, err := loadKpiTrackerDefs()
	if err != nil {
		cfg.logger.Error("Error loading kpiTrackerDefs", "Error", err)
		return nil, err
	}

	kpiTrackerDefs, err = compileRegexStrings(kpiTrackerDefs)
	if err != nil {
		cfg.logger.Error("Error compiling regex strings to Regexp", "Error", err)
		return nil, err
	}

	smartsheetRows := []SmartsheetRow{}

	for _, rfpPackage := range rfpPackages {
		absPath := path.Join(cfg.rfpPackageRootDir, rfpPackage.Name())

		if !rfpPackage.IsDir() {
			cfg.logger.Info("File in the RFP Packages root directory.", "Filename", rfpPackage.Name())
			continue
		}

		rfpProcessedStatus, err := rfpProcessedCompleteCheck(absPath)
		if err != nil {
			cfg.logger.Error("Error checking if RFP Package has been parsed.", "Directory Name", rfpPackage.Name())
			return nil, err
		}

		if rfpProcessedStatus {
			cfg.logger.Info("RFP Package already processed.", "Directory Name", rfpPackage.Name())
			continue
		}
		smartsheetRows, err = cfg.traverseRfpPackage(absPath, kpiTrackerDefs, smartsheetRows)
		if err != nil {
			cfg.logger.Error("Error parsing RFP Package", "Directory Name", rfpPackage.Name())
			continue
		}
	}
	return smartsheetRows, nil
}

func (cfg *apiConfig) traverseRfpPackage(rfpPackage string, kpiTrackerDefs []KpiTrackerDef, smartsheetRows []SmartsheetRow) ([]SmartsheetRow, error) {
	kpiResults := CreateKpiResultForRfpPackage(kpiTrackerDefs)
	stack := []string{rfpPackage}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(current)
		if err != nil {
			cfg.logger.Error("Could not open RFP Package root or sub-directory", "Directory", rfpPackage, "Error", err)
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				dirPath := path.Join(rfpPackage, entry.Name())
				stack = append(stack, dirPath)
				continue
			}

			ext := path.Ext(entry.Name())
			if _, ok := cfg.extMap[ext]; !ok {
				continue
			}

			absPath := path.Join(rfpPackage, entry.Name())

			file, err := os.Open(absPath)
			if err != nil {
				cfg.logger.Error("Error opening file", "Filename", entry.Name(), "Error", err)
				return nil, err
			}
			defer file.Close() //Will not close any files until for loop is complete

			data, err := io.ReadAll(file)
			if err != nil {
				cfg.logger.Error("Error reading file.", "Filename", entry.Name(), "Error", err)
				return nil, err
			}

			kpiResults, err = processFile(data, ext, kpiResults)
			if err != nil {
				cfg.logger.Error("Error parsing file", "Filename", entry.Name(), "Error", err)
				return nil, err
			}
		}
	}
	//func kpiResultsToSmartSheetRow(kpiResults []KpiResult, smartsheetRows []SmartsheetRow) ([]SmartsheetRow, error) {} - Build in rfp_smartsheet_post.go
	// smartsheetRows, err := kpiResultsToSmartSheetRow(kpiResults, smartsheetRows)
	return smartsheetRows, nil
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

func loadKpiTrackerDefs() ([]KpiTrackerDef, error) {
	jsonFile, err := os.Open(kpiTrackerDefPath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	var kpiTrackerDefs []KpiTrackerDef

	err = decoder.Decode(&kpiTrackerDefs)
	if err != nil {
		return nil, err
	}

	if len(kpiTrackerDefs) == 0 {
		return nil, errors.New(".json file does not contain KPI Definition parsing content")
	}

	return kpiTrackerDefs, nil
}

func compileRegexStrings(kpiTrackers []KpiTrackerDef) ([]KpiTrackerDef, error) {
	for i := range kpiTrackers {
		compiled := make([]*regexp.Regexp, 0, len(kpiTrackers[i].RegexStrs))
		for _, pattern := range kpiTrackers[i].RegexStrs {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			compiled = append(compiled, re)
		}
		kpiTrackers[i].Regexps = compiled
	}
	return kpiTrackers, nil
}

func CreateKpiResultForRfpPackage(kpiTrackerDefs []KpiTrackerDef) []KpiResult {
	kpiResults := make([]KpiResult, 0, len(kpiTrackerDefs))

	for i := range kpiTrackerDefs {
		kpiResults = append(kpiResults, KpiResult{
			KpiTrackerDef: &kpiTrackerDefs[i],
			Found:         false,
			Sentence:     "",
		})
	}
	return kpiResults
}
