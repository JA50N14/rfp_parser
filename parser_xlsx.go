package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	sharedStringsFilePath = `/home/jason_macfarlane/tmp`
)

func (cfg *apiConfig) xlsxParser(xlsxData []byte, kpiResults []KpiResult) ([]KpiResult, error) {
	reader := bytes.NewReader(xlsxData)

	zr, err := zip.NewReader(reader, int64(len(xlsxData)))
	if err != nil {
		return kpiResults, fmt.Errorf("Error creating zip reader for xlsx file: %w", err)
	}

	var sharedStrings []string
	if int64(len(xlsxData)) >= int64(104857600) {
		sharedStrings, err = loadSharedStrings(zr)
		if err != nil {
			return kpiResults, err
		}
		kpiResults, err = cfg.parseWithSharedStringsSlice(zr, sharedStrings, kpiResults)
		if err != nil {
			return kpiResults, err
		}
	} else {
		kpiResults, err = cfg.parseWithSharedStringsTmpFile(zr, kpiResults)
		if err != nil {
			return kpiResults, err
		}
	}
	return kpiResults, nil
}

func loadSharedStrings(zr *zip.Reader) ([]string, error) {
	var sharedStrings []string
	var sharedStringsFile *zip.File
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "sharedStrings.xml") {
			sharedStringsFile = f
			break
		}
	}

	if sharedStringsFile == nil {
		return nil, nil
	}

	rc, err := sharedStringsFile.Open()
	if err != nil {
		return nil, fmt.Errorf("Error opening sharedStrings.xml: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var sb strings.Builder
	var inText bool

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error decoding token in sharedStrings.xml: %w", err)
		}

		switch tokElem := tok.(type) {
		case xml.StartElement:
			switch tokElem.Name.Local {
			case "si":
				sb.Reset()
			case "t":
				inText = true
			}
		case xml.CharData:
			if inText {
				sb.Write(tokElem)
			}
		case xml.EndElement:
			switch tokElem.Name.Local {
			case "t":
				inText = false
			case "si":
				sharedStrings = append(sharedStrings, sb.String())
			}
		}
	}
	return sharedStrings, nil
}

func (cfg *apiConfig) parseWithSharedStringsSlice(zr *zip.Reader, sharedStrings []string, kpiResults []KpiResult) ([]KpiResult, error) {
	for _, f := range zr.File {
		if strings.Contains(f.Name, "worksheets/sheet") {
			rc, err := f.Open()
			if err != nil {
				cfg.logger.Error("Error opening worksheet xml file", "Error", err)
				continue
			}
			decoder := xml.NewDecoder(rc)
			var inV bool
			var val string
			var cellType string

			for {
				tok, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					cfg.logger.Error("Error decoding token in worksheet xml file", "Error", err)
					continue
				}

				switch tokElem := tok.(type) {
				case xml.StartElement:
					if tokElem.Name.Local == "c" {
						for _, attr := range tokElem.Attr {
							if attr.Name.Local == "t" {
								cellType = attr.Value
								break
							}
						}
					}
					if tokElem.Name.Local == "v" {
						inV = true
						val = ""
					}
				case xml.CharData:
					if inV {
						val += string(tokElem)
					}
				case xml.EndElement:
					if tokElem.Name.Local == "v" {
						inV = false
						if cellType == "s" {
							idx, _ := strconv.Atoi(val)
							if idx < 0 || idx >= len(sharedStrings) {
								fmt.Printf("Skipping invalid shared string index %d (len=%d)\n", idx, len(sharedStrings))
								continue
							}
							text := sharedStrings[idx]
							kpiResults = scanTextWithRegex(text, kpiResults)
						} else {
							kpiResults = scanTextWithRegex(val, kpiResults)
						}
					}
				}
			}
			rc.Close()
		}
	}
	return kpiResults, nil
}

func (cfg *apiConfig) parseWithSharedStringsTmpFile(zr *zip.Reader, kpiResults []KpiResult) ([]KpiResult, error) {
	var sharedStringsFile *zip.File
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "sharedStrings.xml") {
			sharedStringsFile = f
			break
		}
	}

	var ssFile *os.File
	var ssOffsets []int64
	ssIndex := 0
	hasSharedStrings := sharedStringsFile != nil

	if hasSharedStrings {
		rc, err := sharedStringsFile.Open()
		if err != nil {
			return kpiResults, fmt.Errorf("Error opening sharedStrings.xml: %w", err)
		}

		decoder := xml.NewDecoder(rc)

		var sb strings.Builder
		var inText bool
		ssFile, err = os.CreateTemp(sharedStringsFilePath, "sharedStrings")
		if err != nil {
			return kpiResults, fmt.Errorf("Error created sharedStrings temporary file: %w", err)
		}
		defer os.Remove(ssFile.Name())

		//Fill up ssFile and ssOffsets
		for {
			tok, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				return kpiResults, fmt.Errorf("Error decoding token in sharedStrings.xml: %w", err)
			}

			switch tokElem := tok.(type) {
			case xml.StartElement:
				switch tokElem.Name.Local {
				case "si":
					sb.Reset()
				case "t":
					inText = true
				}
			case xml.CharData:
				if inText {
					sb.Write(tokElem)
				}
			case xml.EndElement:
				switch tokElem.Name.Local {
				case "t":
					inText = false
				case "si":
					offset, _ := ssFile.Seek(0, io.SeekCurrent)
					ssOffsets = append(ssOffsets, offset)
					fmt.Fprintf(ssFile, "%d|%s\n", ssIndex, sb.String())
					ssIndex++
				}
			}
		}
		rc.Close()
	}

	//Parse worksheets
	for _, f := range zr.File {
		if strings.Contains(f.Name, "worksheets/sheet") {
			rc, err := f.Open()
			if err != nil {
				return kpiResults, err
			}
			decoder := xml.NewDecoder(rc)
			var inV bool
			var val string
			var cellType string

			for {
				tok, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					cfg.logger.Error("Error decoding token in worksheet xml file", "Error", err)
					continue
				}

				switch tokElem := tok.(type) {
				case xml.StartElement:
					if tokElem.Name.Local == "c" {
						for _, attr := range tokElem.Attr {
							if attr.Name.Local == "t" {
								cellType = attr.Value
							}
						}
					}
					if tokElem.Name.Local == "v" {
						inV = true
						val = ""
					}
				case xml.CharData:
					if inV {
						val += string(tokElem)
					}
				case xml.EndElement:
					if tokElem.Name.Local == "v" {
						inV = false
						if cellType == "s" {
							valIdx, _ := strconv.Atoi(val)
							if valIdx < 0 || valIdx > ssIndex {
								fmt.Printf("Skipping invalid shared string index: %d (len=%d)\n", valIdx, ssIndex)
								continue
							}
							_, err = ssFile.Seek(ssOffsets[valIdx], io.SeekStart)
							if err != nil {
								return kpiResults, fmt.Errorf("Error retrieving text from sharedStrings tmp file: %w", err)
							}
							reader := bufio.NewReader(ssFile)
							line, err := reader.ReadString('\n')
							if err != nil && err != io.EOF {
								return kpiResults, fmt.Errorf("Error reading sharedString text from sharedString file: %w", err)
							}
							parts := strings.SplitN(line, "|", 2)
							if len(parts) != 2 {
								return kpiResults, fmt.Errorf("Invalid text format in sharedString tmp file: %w", err)
							}
							kpiResults = scanTextWithRegex(strings.TrimSpace(parts[1]), kpiResults)
						} else {
							kpiResults = scanTextWithRegex(val, kpiResults)
						}
					}
				}
			}
			rc.Close()
		}
	}
	return kpiResults, nil
}

//<c> - cell element
// r = cell reference (i.e. A1, B1, etc.)
// t = cell type ("s" means "shared string", i.e. text stored in sharedStrings.xml)
//<v> = value (an integer index if t="s", or a literal value otherwise)

//xl/sharedStrings.xml
//<si> = Shared string item
//<t> = Actual text
//If a cell has t="s" and <v>1</v>, that means the cell text is in sharedStrings.xml file as the second <si>
