package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"code.sajari.com/docconv"
	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
)

const (
	docxExt = ".docx"
	docExt  = ".doc"
	xlsxExt = ".xlsx"
	xlsExt  = ".xls"
)

type cleanupRule struct {
	re   *regexp.Regexp
	repl string
}

type cellData struct {
	sheet string
	row   int
	col   int
	value string
}

var cleanupRules = []cleanupRule{
	{regexp.MustCompile(`\n[©•-]\n`), " "},
	{regexp.MustCompile(`[ \t]+`), " "},
	{regexp.MustCompile(`([A-Za-z])\n([a-z])`), "$1$2"},
	{regexp.MustCompile(`([A-Za-z]) \n\n([A-Za-z])`), "$1 $2"},
	{regexp.MustCompile(`([A-Za-z]) \n([A-Za-z])`), "$1 $2"},
	{regexp.MustCompile(`\n \n`), " "},
	{regexp.MustCompile(`\.\n \n`), ". "},
	{regexp.MustCompile(`\n\n+`), "\n"},
	{regexp.MustCompile(`(?m)^[-•]\s*`), ""},
}

var sentenceRule = regexp.MustCompile(`\. [A-Z]`)

func processFile(filePath string, fileExt string, kpiResults []KpiResult) ([]KpiResult, error) {
	var text string
	wbData := make(map[string][][]string)
	var err error

	switch fileExt {
	case docxExt:
		text, err = extractTextFromDocx(filePath)
	case docExt:
		text, err = extractTextFromDoc(filePath)
	case xlsxExt:
		wbData, err = extractTextFromXlsx(filePath)
	case xlsExt:
		wbData, err = extractTextFromXls(filePath)
	default:
		return kpiResults, fmt.Errorf("File cannot be parsed due to incorrect file type: %s", fileExt)
	}

	if err != nil || (text == "" && len(wbData) == 0) {
		return kpiResults, err
	}

	if fileExt == docxExt || fileExt == docExt {
		kpiResults = textParser(text, kpiResults)
	}

	if fileExt == xlsxExt || fileExt == xlsExt {
		kpiResults = xlsxAndXlsParser(wbData, kpiResults)
	}

	return kpiResults, nil
}

func extractTextFromDocx(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Error opening .docx file: %w", err)
	}
	defer f.Close()

	text, _, err := docconv.ConvertDocx(f)
	if err != nil {
		return "", fmt.Errorf("Error parsing .docx file: %w", err)
	}
	return text, nil
}

func extractTextFromDoc(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Error opening .doc file: %w", err)
	}
	defer f.Close()

	text, _, err := docconv.ConvertDoc(f)
	if err != nil {
		return "", fmt.Errorf("Error parsing .doc file: %w", err)
	}
	return text, nil
}

func extractTextFromXlsx(filePath string) (map[string][][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening .xlsx file: %w", err)
	}
	defer f.Close()

	wb, err := excelize.OpenReader(f)
	if err != nil {
		return nil, fmt.Errorf("Error opening .xlsx reader: %w", err)
	}

	wbData := make(map[string][][]string)
	for _, sheetName := range wb.GetSheetList() {
		rows, err := wb.GetRows(sheetName)
		if err != nil {
			return nil, err
		}
		wbData[sheetName] = rows
	}
	return wbData, nil
}

func extractTextFromXls(filePath string) (map[string][][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening .xls file: %w", err)
	}
	defer f.Close()

	wb, err := xls.OpenReader(f, "utf-8")
	if err != nil {
		return nil, fmt.Errorf("Error opening .xls reader: %w", err)
	}

	wbData := make(map[string][][]string)
	for i := 0; 1 < int(wb.NumSheets()); i++ {
		sheet := wb.GetSheet(i)
		if sheet == nil {
			continue
		}

		var sheetRows [][]string
		for r := 0; r <= int(sheet.MaxRow); r++ {
			row := sheet.Row(r)
			var rowCells []string
			for c := 0; c < row.LastCol(); c++ {
				rowCells = append(rowCells, row.Col(c))
			}
			sheetRows = append(sheetRows, rowCells)
		}
		wbData[sheet.Name] = sheetRows
	}
	return wbData, nil
}

func textParser(text string, kpiResults []KpiResult) []KpiResult {
	text = cleanText(text)
	textSlice := strings.Split(text, "\n")

	for _, item := range textSlice {
		for i, kpiResult := range kpiResults {
			for _, re := range kpiResult.Regexps {
				if re.Match([]byte(item)) {
					sentence := extractSentence(item, re, sentenceRule)
					if sentence == "" {
						continue
					}
					kpiResults[i].Sentence = sentence
					kpiResults[i].Found = true
					break
				}
			}
		}
	}

	kpiResults = removeKpiResultsNotFound(kpiResults)
	return kpiResults
}

func cleanText(text string) string {
	for _, rule := range cleanupRules {
		text = rule.re.ReplaceAllString(text, rule.repl)
	}
	return text
}

func extractSentence(text string, target *regexp.Regexp, boundary *regexp.Regexp) string {
	targetLoc := target.FindIndex([]byte(text))
	if targetLoc == nil {
		return ""
	}

	leftBounds := boundary.FindAllStringIndex(text[0:targetLoc[0]], -1)
	leftIdx := 0
	if leftBounds != nil {
		leftBoundsLen := len(leftBounds) - 1
		leftIdx = leftBounds[leftBoundsLen][0] + 2
	}

	rightBound := boundary.FindStringIndex(text[targetLoc[1]:])
	rightIdx := len(text)
	if rightBound != nil {
		rightIdx = targetLoc[1] + rightBound[1] - 2
	}
	return text[leftIdx:rightIdx]
}

func xlsxAndXlsParser(wbData map[string][][]string, kpiResults []KpiResult) []KpiResult {
	allData := flattenXlsxAndXlsData(wbData)
	for _, cell := range allData {
		for i, kpiResult := range kpiResults {
			for _, re := range kpiResult.Regexps {
				if re.Match([]byte(cell.value)) {
					sentence := extractSentence(cell.value, re, sentenceRule)
					if sentence == "" {
						continue
					}
					kpiResults[i].Sentence = sentence
					kpiResults[i].Found = true
				}
			}
		}
	}
	kpiResults = removeKpiResultsNotFound(kpiResults)
	return kpiResults
}

func flattenXlsxAndXlsData(wbData map[string][][]string) []cellData {
	var allData []cellData
	for sheetName, sheetContent := range wbData {
		for rowNum, row := range sheetContent {
			for colNum, cell := range row {
				cd := cellData{
					sheet: sheetName,
					row:   rowNum + 1,
					col:   colNum + 1,
					value: cell,
				}
				allData = append(allData, cd)
			}
		}
	}
	return allData
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
