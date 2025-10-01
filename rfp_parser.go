package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"code.sajari.com/docconv"
	"github.com/extrame/xls"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

const (
	pdfExt  = ".pdf"
	docxExt = ".docx"
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
	sheetsData := make(map[string][][]string)
	var err error

	switch fileExt {
	case pdfExt:
		text, err = extractTextFromPdf(filePath)
	case docxExt:
		text, err = extractTextFromDocx(filePath)
	case xlsxExt:
		sheetsData, err = extractTextFromXlsx(filePath)
	case xlsExt:
		sheetsData, err = extractTextFromXls(filePath)
	default:
		return kpiResults, fmt.Errorf("File cannot be parsed due to incorrect file type: %s", fileExt)
	}

	if err != nil || (text == "" && len(sheetsData) == 0) {
		return kpiResults, err
	}

	if fileExt == pdfExt || fileExt == docxExt {
		kpiResults = pdfAndDocxParser(text, kpiResults)
	}

	if fileExt == xlsxExt || fileExt == xlsExt {
		kpiResults = xlsxAndXlsParser(sheetsData, kpiResults)
	}

	return kpiResults, nil
}

func extractTextFromPdf(filePath string) (string, error) {
	pdf.DebugOn = true

	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Could not open PDF: %w", err)
	}
	defer f.Close()

	var buff bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		if strings.Contains(err.Error(), "unexpected") {
			return "", fmt.Errorf("skipping due to unexpected characters in PDF: %w", err)
		}
		if strings.Contains(err.Error(), "invalid header") {
			return "", fmt.Errorf("skipping due to invalid compressed stream in PDF: %w", err)
		}
		return "", fmt.Errorf("skipping due to unable to get text from PDF: %w", err)
	}

	buff.ReadFrom(b)
	return buff.String(), nil
}

func extractTextFromDocx(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Could not open docx: %w", err)
	}
	defer f.Close()

	text, _, err := docconv.ConvertDocx(f)
	if err != nil {
		return "", err
	}
	return text, nil
}

func extractTextFromXlsx(filePath string) (map[string][][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Could not open xlsx: %w", err)
	}
	defer f.Close()

	wb, err := excelize.OpenReader(f)
	if err != nil {
		return nil, err
	}

	sheetsData := make(map[string][][]string)
	for _, sheetName := range wb.GetSheetList() {
		rows, err := wb.GetRows(sheetName)
		if err != nil {
			return nil, err
		}
		sheetsData[sheetName] = rows
	}
	return sheetsData, nil
}

func extractTextFromXls(filePath string) (map[string][][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Could not open xls: %w", err)
	}

	wb, err := xls.OpenReader(f, "utf-8")
	if err != nil {
		return nil, err
	}

	sheetsData := make(map[string][][]string)

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
		sheetsData[sheet.Name] = sheetRows
	}
	return sheetsData, nil
}

func pdfAndDocxParser(text string, kpiResults []KpiResult) []KpiResult {
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

func cleanText(text string) string {
	for _, rule := range cleanupRules {
		text = rule.re.ReplaceAllString(text, rule.repl)
	}
	return text
}

func xlsxAndXlsParser(sheetsData map[string][][]string, kpiResults []KpiResult) []KpiResult {
	allData := flattenXlsxAndXlsData(sheetsData)
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

func flattenXlsxAndXlsData(sheetsData map[string][][]string) []cellData {
	var allData []cellData
	for sheetName, sheetContent := range sheetsData {
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
