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

func resultsToSmartsheetRows(allResults []PackageResult) (AddRowsRequest, error) {
	smartsheetRows := AddRowsRequest {
		ToTop: true,
	}

	for _, package := range allResults {
		var row Row
		for _, result := range package.Results {
			cellPackageName := Cell {
				ColumnId: 1,
				Value: allResults.PackageName,
			}
			cellKpi := Cell {
				ColumnId: result.ColumnID,
				Value: result.Name
			}
			cellCategory := Cell {
				ColumnId: result.
			}

		}
	}

}

//create a AddRowsRequest struct
//Loop through each PackageResult
//Create a Row struct
//Loop through each Result
//create a Cell struct
//Populate Cell with ColumnID & PackageName, Name, Category, Sentence
