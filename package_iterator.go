package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

type KpiTrackerDef struct {
	Name      string           `json:"name"`
	Category  string           `json:"category"`
	Regexps   []*regexp.Regexp `json:"-"`
	RegexStrs []string         `json:"regexps"` //temporary holder
}

type KpiResult struct {
	*KpiTrackerDef
	Found    bool
	Sentence string
}

type PackageResult struct {
	PackageName string
	DateParsed  string
	Results     []KpiResult
}

const (
	kpiTrackerDefPath = "./kpiTracker.json"
	docxExt           = ".docx"
	xlsxExt           = ".xlsx"
)

func (cfg *apiConfig) traverseRfpPackages() ([]PackageResult, error) {
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

	var allResults []PackageResult

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

		packageResult, err := cfg.traverseRfpPackage(absPath, kpiTrackerDefs)
		if err != nil {
			cfg.logger.Error("Error parsing RFP Package", "Directory Name", rfpPackage.Name())
			continue
		}
		allResults = append(allResults, packageResult)
	}
	return allResults, nil
}

func (cfg *apiConfig) traverseRfpPackage(rfpPackage string, kpiTrackerDefs []KpiTrackerDef) (PackageResult, error) {
	kpiResults := CreateKpiResultForRfpPackage(kpiTrackerDefs)
	stack := []string{rfpPackage}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(current)
		if err != nil {
			cfg.logger.Error("Could not open RFP Package root directory or sub-directory", "Directory", rfpPackage, "Error", err)
			return PackageResult{}, err
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

			filePath := path.Join(rfpPackage, entry.Name())

			data, err := fileToBytes(filePath)
			if err != nil {
				cfg.logger.Error("Error opening file to read into bytes", "Error", err)
				continue
			}

			switch path.Ext(entry.Name()) {
			case docxExt:
				kpiResults, err = docxParser(data, kpiResults)
			case xlsxExt:
				kpiResults, err = cfg.xlsxParser(data, kpiResults)
			default:
				continue
			}
			if err != nil {
				cfg.logger.Error("Error parsing file", "Filename", entry.Name(), "Package", rfpPackage, "Error", err)
				continue
			}
		}
	}
	kpiResults = removeKpiResultsNotFound(kpiResults)
	packageResult := PackageResult{
		PackageName: filepath.Base(rfpPackage),
		DateParsed:  time.Now().Format("2006-01-02"),
		Results:     kpiResults,
	}
	return packageResult, nil
}

func fileToBytes(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func rfpProcessedCompleteCheck(rfpRootDir string) (bool, error) {
	rfpPackage, err := os.ReadDir(rfpRootDir)
	if err != nil {
		return false, fmt.Errorf("Could not open RFP Package: %w", err)
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
			Sentence:      "",
		})
	}
	return kpiResults
}

func removeKpiResultsNotFound(kpiResults []KpiResult) []KpiResult {
	var kpiResultsFound []KpiResult
	for _, result := range kpiResults {
		if result.Found == true {
			kpiResultsFound = append(kpiResultsFound, result)
		}
	}
	return kpiResultsFound

}
