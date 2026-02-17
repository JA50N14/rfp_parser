package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Cell struct {
	ColumnId int64  `json:"columnId"`
	Value    string `json:"value"`
}

type Row struct {
	ToTop bool   `json:"toTop"`
	Cells []Cell `json:"cells"`
}

// const (
// 	colDateParsed  int64 = 5732040604077956
// 	colYear int64 = 6705789409120132
// 	colBusinessUnit int64 = 4453989595434884
// 	colPackageName int64 = 3480240790392708
// 	colKpiName     int64 = 7983840417763204
// 	colKpiCategory int64 = 665491023286148
// 	colKpiSentence int64 = 4756959379804036
// )

//Test Smartsheet Column ID's
const (
	colDateParsed     int64 = 4915691982114692
	colYear           int64 = 2663892168429444
	colBusinessUnit   int64 = 7167491795799940
	colDivision       int64 = 1537992261586820
	colPackageName int64 = 6041591888957316
	colKpiName        int64 = 3789792075272068
	colKpiCategory    int64 = 8293391702642564
	colKpiSentence     int64 = 975042308165508
)


func (cfg *apiConfig) postRequestSmartsheets(smartsheetRows []Row) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	reqBody := &bytes.Buffer{}
	encoder := json.NewEncoder(reqBody)
	err := encoder.Encode(smartsheetRows)
	if err != nil {
		cfg.logger.Error("Error encoding smartsheetRows to json", "Error", err)
		return err
	}

	req, err := http.NewRequest("POST", cfg.smartsheetUrl, reqBody)
	if err != nil {
		cfg.logger.Error("Error creating request", "Error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.bearerTokenSmartsheet)

	resp, err := client.Do(req)
	if err != nil {
		cfg.logger.Error("Error making post request to Smartsheets", "Error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		cfg.logger.Error("Non-2xx status code from Smartsheet", "Status", resp.StatusCode, "Body", string(body))
		return fmt.Errorf("Smartsheet return status %d", resp.StatusCode)
	}
	return nil
}

func resultsToSmartsheetRows(allResults []PackageResult) []Row {
	var smartsheetRows []Row

	for _, rfpPackage := range allResults {
		for _, result := range rfpPackage.Results {
			row := Row{
				ToTop: true,
				Cells: []Cell{
					{
						ColumnId: colDateParsed,
						Value:    rfpPackage.DateParsed,
					},
					{
						ColumnId: colYear,
						Value: rfpPackage.PackageYear,
					},
					{
						ColumnId: colBusinessUnit,
						Value: rfpPackage.BusinessUnit,
					},
					{
						ColumnId: colDivision,
						Value: rfpPackage.Division,
					},
					{
						ColumnId: colPackageName,
						Value:    rfpPackage.PackageName,
					},
					{
						ColumnId: colKpiName,
						Value:    fmt.Sprintf("%v", result.Name),
					},
					{
						ColumnId: colKpiCategory,
						Value:    fmt.Sprintf("%v", result.Category),
					},
					{
						ColumnId: colKpiSentence,
						Value:    fmt.Sprintf("%v", result.Sentence),
					},
				},
			}
			smartsheetRows = append(smartsheetRows, row)
		}
	}
	return smartsheetRows
}
