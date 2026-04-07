package walk

import (
	"os"
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
	pdfExt  = ".pdf"
)

func ProcessRFPPackage(pkg graph.Package, path WalkPath, walkCtx *WalkContext) (PkgResult, error) {
	kpiResults := parser.CreatePkgResultForRFPPackage(walkCtx.KPIDefs)

	items, err := graph.GetItemSubDirs(pkg.ID, walkCtx.Ctx, walkCtx.Cfg)
	if err != nil {
		return PkgResult{}, err
	}

	for _, item := range items {
		if err := walkRFPPackage(item, pkg, path, kpiResults, walkCtx); err != nil {
			return PkgResult{}, err
		}
	}

	kpiResults = parser.RemoveKPIResultsNotFound(kpiResults)

	pkgResult := PkgResult{
		PackageName:  pkg.Name,
		DateParsed:   time.Now().Format("2006-01-02"),
		Year:         path.Year,
		BusinessUnit: path.BusinessUnit,
		Division:     path.Division,
		KPIResults:   kpiResults,
	}

	return pkgResult, nil
}

func walkRFPPackage(item graph.Item, pkg graph.Package, path WalkPath, kpiResults []parser.KPIResult, walkCtx *WalkContext) error {
	ext := filepath.Ext(item.Name)

	switch ext {
	case docxExt:
		f, err := graph.GetFile(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}
		defer f.Close()
		defer os.Remove(f.Name())

		info, err := f.Stat()
		if err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}

		if err := parser.DocxParser(f, info.Size(), kpiResults); err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}
		return nil

	case xlsxExt:
		f, err := graph.GetFile(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}
		defer f.Close()
		defer os.Remove(f.Name())

		info, err := f.Stat()
		if err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}

		if err := parser.XlsxParser(f, info.Size(), kpiResults); err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}
		return nil

	case pdfExt:
		f, err := graph.GetFile(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}
		defer f.Close()
		defer os.Remove(f.Name())

		if err := parser.PdfParser(walkCtx.Ctx, f, kpiResults); err != nil {
			walkCtx.Cfg.Logger.Warn("Unable to process file", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "File Name", item.Name, "error", err)
			return nil
		}
		return nil

	case "":
		childItems, err := graph.GetItemSubDirs(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}

		for _, childItem := range childItems {
			if err := walkRFPPackage(childItem, pkg, path, kpiResults, walkCtx); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}
