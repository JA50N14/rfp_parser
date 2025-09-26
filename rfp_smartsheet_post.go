package main

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
	colDateParsed = 1
	colPackageName = 2
	colKpiName = 3
	colKpiCategory = 4
	colKpiSentence = 5
)


func resultsToSmartsheetRows(allResults []PackageResult) (AddRowsRequest, error) {
	smartsheetRows := AddRowsRequest {
		ToTop: true,
	}

	for _, package := range allResults {
		for _, result := range package.Results {
			var row Row
			
			row := Row {
				cellParseDate := Cell {
				ColumnId: colDateParsed,
				Value: package.DateParsed
				},
				cellPackageName := Cell {
					ColumnId: colPackageName,
					Value: package.PackageName,
				},

			cellKpiFound := Cell {
				ColumnId: colKpiName,
				Value: result.Name,
			}

			cellCategory := Cell {
				ColumnId: colKpiCategory,
				Value: result.Category,
			}

			cellKpiSentence := Cell {
				ColumnId: colKpiSentence,
				Value: result.Sentence,
			}
			}


			
		}
	}
}

//create a AddRowsRequest struct
//Loop through each PackageResult
	//Loop through each Result
		//Create a Row struct
		//create a Cell struct
		//Populate Cell with ColumnID & PackageName, Name, Category, Sentence
