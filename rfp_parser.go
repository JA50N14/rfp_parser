package main

import (
	"bytes"
	"fmt"
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

func processFile(data []byte, fileExt string, kpiResults []KpiResult) ([]KpiResult, error) {
	var text string
	sheetsData := make(map[string][][]string)
	var err error

	switch fileExt {
	case pdfExt:
		text, err = extractTextFromPdf(data)
	case docxExt:
		text, err = extractTextFromDocx(data)
	case xlsxExt:
		sheetsData, err = extractTextFromXlsx(data)
	case xlsExt:
		sheetsData, err = extractTextFromXls(data)
	}

	if err != nil {
		return nil, err
	}

	if fileExt == pdfExt || fileExt == docxExt {
		kpiResults = pdfAndDocxParser(text, kpiResults)
	}

	if fileExt == xlsxExt || fileExt == xlsExt {
		kpiResults, err = xlsxAndXlsParser(sheetsData, kpiResults)
		if err != nil {
			return nil, err
		}
	}

	return kpiResults, nil
}

func extractTextFromPdf(data []byte) (string, error) {
	pdf.DebugOn = true

	pdfReader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := pdfReader.GetPlainText()
	if err != nil {
		return "", err
	}

	buf.ReadFrom(b)
	text := buf.String()
	return text, nil
}

func extractTextFromDocx(data []byte) (string, error) {
	text, _, err := docconv.ConvertDocx(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	return text, nil
}

func extractTextFromXlsx(data []byte) (map[string][][]string, error) {
	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sheetsData := make(map[string][][]string)
	for _, sheetName := range file.GetSheetList() {
		rows, err := file.GetRows(sheetName)
		if err != nil {
			return nil, err
		}
		sheetsData[sheetName] = rows
	}
	return sheetsData, nil
}

func extractTextFromXls(data []byte) (map[string][][]string, error) {
	file, err := xls.OpenReader(bytes.NewReader(data), "utf-8")
	if err != nil {
		return nil, err
	}

	sheetsData := make(map[string][][]string)

	for i := 0; 1 < int(file.NumSheets()); i++ {
		sheet := file.GetSheet(i)
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

	//TEST CODE:
	// os.WriteFile("output.txt", []byte(text), 0644)
	// fmt.Println(strings.Join(textSlice, ","))
	// fmt.Println("Single Element in textSlice:")
	// fmt.Println(textSlice[3])
	// fmt.Println(len(textSlice))
	//-------------------

	for _, item := range textSlice {
		for idx, kpiResult := range kpiResults {
			for i, re := range kpiResult.Regexps {
				if kpiResult.Found == true {
					continue
				}
				if re.Match([]byte(item)) {
					sentence := extractSentence(item, kpiResult.Regexps[i], sentenceRule)
					if sentence == "" {
						continue
					}
					kpiResults[idx].Sentence = sentence
					kpiResults[idx].Found = true
				}
			}
		}
	}
	//FOR TESTING
	for _, kpiResult := range kpiResults {
		fmt.Printf(">kpiResults - Name: %v / Found: %v / Sentence: %s\n", kpiResult.Name, kpiResult.Found, kpiResult.Sentence)
	}
	//-------------
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

func xlsxAndXlsParser(sheetsData map[string][][]string, kpiResults []KpiResult) ([]KpiResult, error) {
	return kpiResults, nil
}
