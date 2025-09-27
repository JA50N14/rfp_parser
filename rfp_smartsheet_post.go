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
	ColumnId int    `json:"columnId"`
	Value    string `json:"value"`
}

type Row struct {
	Cells []Cell `json:"cells"`
}

type AddRowsRequest struct {
	ToTop bool  `json:"toTop"`
	Rows  []Row `json:"rows"`
}

const (
	colDateParsed  = 5732040604077956
	colPackageName = 3480240790392708
	colKpiName     = 7983840417763204
	colKpiCategory = 665491023286148
	colKpiSentence = 4756959379804036
)

func (cfg *apiConfig) postRequestSmartsheets(smartsheetRows AddRowsRequest) error {
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

func resultsToSmartsheetRows(allResults []PackageResult) AddRowsRequest {
	smartsheetRows := AddRowsRequest{
		ToTop: false,
	}

	for _, rfpPackage := range allResults {
		for _, result := range rfpPackage.Results {
			row := Row{
				Cells: []Cell{
					{
						ColumnId: colDateParsed,
						Value:    rfpPackage.DateParsed,
					},
					{
						ColumnId: colPackageName,
						Value:    rfpPackage.PackageName,
					},
					{
						ColumnId: colKpiName,
						Value:    result.Name,
					},
					{
						ColumnId: colKpiCategory,
						Value:    result.Category,
					},
					{
						ColumnId: colKpiSentence,
						Value:    result.Sentence,
					},
				},
			}
			smartsheetRows.Rows = append(smartsheetRows.Rows, row)
		}
	}
	return smartsheetRows
}
