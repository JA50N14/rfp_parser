package walk

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/JA50N14/rfp_parser/graph"
	"github.com/JA50N14/rfp_parser/parser"
)

type PkgResult struct {
	PackageName  string
	DateParsed   string
	Year         string
	BusinessUnit string
	Division     string
	KPIResults   []parser.KPIResult
}

const (
	docxExt = ".docx"
	xlsxExt = ".xlsx"
)

func ProcessRFPPackage(pkg graph.Package, path WalkPath, walkCtx *WalkContext) (PkgResult, error) {
	kpiResults := parser.CreatePkgResultForRFPPackage(walkCtx.KPIDefs)

	items, err := graph.GetItemSubDirs(pkg.ID, walkCtx.Ctx, walkCtx.Cfg)
	if err != nil {
		return PkgResult{}, err
	}

	for _, item := range items {
		if err := walkRFPPackage(item, kpiResults, walkCtx); err != nil {
			return PkgResult{}, err
		}
	}

	kpiResults = parser.RemoveKPIResultsNotFound(kpiResults)

	pkgResult := PkgResult{
		PackageName:  pkg.ID,
		DateParsed:   time.Now().Format("2026-01-02"),
		Year:         path.Year,
		BusinessUnit: path.BusinessUnit,
		Division:     path.Division,
		KPIResults:   kpiResults,
	}

	return pkgResult, nil
}

func walkRFPPackage(item graph.Item, kpiResults []parser.KPIResult, walkCtx *WalkContext) error {
	ext := filepath.Ext(item.Name)

	switch ext {
	case docxExt:
		f, err := graph.GetFile(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("unable to get file stats: %w", err)
		}

		if err := parser.DocxParser(f, info.Size(), kpiResults); err != nil {
			return fmt.Errorf("error parsing docx file: Item ID: %s, Name: %s, error: %w", item.ID, item.Name, err)
		}
		return nil

	case xlsxExt:
		f, err := graph.GetFile(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("unable to get file stats: %w", err)
		}

		if err := parser.XlsxParser(f, info.Size(), kpiResults); err != nil {
			return fmt.Errorf("error parsing xlsx file: Item ID: %s, Name: %s, error: %w", item.ID, item.Name, err)
		}
		return nil

	case "":
		childItems, err := graph.GetItemSubDirs(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}

		for _, childItem := range childItems {
			if err := walkRFPPackage(childItem, kpiResults, walkCtx); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}


