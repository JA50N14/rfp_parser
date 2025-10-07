package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func docxParser(docxData []byte, kpiResults []KpiResult) ([]KpiResult, error) {
	reader := bytes.NewReader(docxData)

	zr, err := zip.NewReader(reader, int64(len(docxData)))
	if err != nil {
		return kpiResults, fmt.Errorf("Error creating zip reader: %w", err)
	}

	var documentFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			documentFile = f
			break
		}
	}

	if documentFile == nil {
		return kpiResults, fmt.Errorf("word/document.xml not found")
	}

	rc, err := documentFile.Open()
	if err != nil {
		return kpiResults, fmt.Errorf("Error opening word/document.xml: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var (
		inText      bool
		currentText string
		output      strings.Builder
	)

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return kpiResults, err
		}

		switch tokElem := tok.(type) {
		case xml.StartElement:
			if tokElem.Name.Local == "t" {
				inText = true
				currentText = ""
			}
		case xml.CharData:
			if inText {
				currentText += string(tokElem)
			}
		case xml.EndElement:
			if tokElem.Name.Local == "t" {
				output.WriteString(currentText)
				inText = false
			}
			if tokElem.Name.Local == "p" {
				output.WriteString("\n")
				kpiResults = scanTextWithRegex(output.String(), kpiResults)
				output.Reset()
			}
		}
	}
	return kpiResults, nil
}


//word/document.xml file:
//<w:p> = paragraph
//<w:r> = run (chunk of text with consistent formatting)
//<w:t> = the actual text node

