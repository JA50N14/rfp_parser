package walk

import (
	"fmt"
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
// 	colDateParsed     int64 = 5732040604077956
// 	colYear           int64 = 6705789409120132
// 	colBusinessUnit   int64 = 4453989595434884
// 	colDivision       int64 = 336189612314500
// 	colRFPPackageName int64 = 3480240790392708
// 	colKPIName        int64 = 7983840417763204
// 	colKPICategory    int64 = 665491023286148
// 	colKPIContext     int64 = 4756959379804036
// )

//Test Smartsheet Column ID's
const (
	colDateParsed     int64 = 4915691982114692
	colYear           int64 = 2663892168429444
	colBusinessUnit   int64 = 7167491795799940
	colDivision       int64 = 1537992261586820
	colRFPPackageName int64 = 6041591888957316
	colKPIName        int64 = 3789792075272068
	colKPICategory    int64 = 8293391702642564
	colKPIContext     int64 = 975042308165508
)


func prepareResultsForSmartsheetRows(result PkgResult) []Row {
	var smartsheetRows []Row

	for _, kpiResult := range result.KPIResults {
		row := Row{
			ToTop: true,
			Cells: []Cell{
				{
					ColumnId: colDateParsed,
					Value:    result.DateParsed,
				},
				{
					ColumnId: colYear,
					Value:    result.Year,
				},
				{
					ColumnId: colBusinessUnit,
					Value:    result.BusinessUnit,
				},
				{
					ColumnId: colDivision,
					Value:    result.Division,
				},
				{
					ColumnId: colRFPPackageName,
					Value:    result.PackageName,
				},
				{
					ColumnId: colKPIName,
					Value:    fmt.Sprintf("%v", kpiResult.KPIDef.Name),
				},
				{
					ColumnId: colKPICategory,
					Value:    fmt.Sprintf("%v", kpiResult.KPIDef.Category),
				},
				{
					ColumnId: colKPIContext,
					Value:    fmt.Sprintf("%v", kpiResult.Sentence),
				},
			},
		}
		smartsheetRows = append(smartsheetRows, row)
	}

	return smartsheetRows
}