package walk

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/graph"
	"github.com/JA50N14/rfp_parser/parser"
)

type WalkContext struct {
	Cfg     *config.ApiConfig
	Ctx     context.Context
	Now     time.Time
	KPIDefs []parser.KPIDefinition
	Results *[]PkgResult
}

type WalkPath struct {
	Year         string
	BusinessUnit string
	Division     string
}

type Level int

const (
	LevelYear Level = iota
	LevelBusinessUnit
	LevelDivision
)

func WalkDocLibrary(ctx context.Context, cfg *config.ApiConfig) ([]PkgResult, error) {
	kpiDefs, err := parser.LoadKPIDefinitions()
	if err != nil {
		return nil, err
	}

	kpiDefs, err = parser.CompileRegexStrings(kpiDefs)
	if err != nil {
		return nil, err
	}

	results := make([]PkgResult, 0)

	walkCtx := &WalkContext{
		Cfg:     cfg,
		Ctx:     ctx,
		Now:     time.Now(),
		KPIDefs: kpiDefs,
		Results: &results,
	}

	rootDirs, err := graph.GetRootDirs(walkCtx.Ctx, walkCtx.Cfg)
	if err != nil {
		return nil, err
	}

	for _, dir := range rootDirs {
		ok := isValidYear(dir.Name)
		if !ok {
			walkCtx.Cfg.Logger.Info("invalid year directory at root level", "year", dir.Name)
			continue
		}

		path := WalkPath{
			Year: dir.Name,
		}

		if err := Walk(dir, LevelYear, path, walkCtx); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func Walk(item graph.Item, level Level, path WalkPath, walkCtx *WalkContext) error {
	switch level {
	case LevelYear:
		items, err := graph.GetItemSubDirs(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}

		for _, item := range items {
			nextPath := path
			nextPath.BusinessUnit = item.Name
			Walk(item, LevelBusinessUnit, nextPath, walkCtx)
		}
	case LevelBusinessUnit:
		items, err := graph.GetItemSubDirs(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}
		for _, item := range items {
			nextPath := path
			nextPath.Division = item.Name
			Walk(item, LevelDivision, nextPath, walkCtx)
		}
	case LevelDivision:
		pkgs, err := graph.GetItemSubDirsWithMetadata(item.ID, walkCtx.Ctx, walkCtx.Cfg)
		if err != nil {
			return err
		}

		pkgs, err = removeProcessedPackages(pkgs)
		if err != nil {
			return err
		}

		for _, pkg := range pkgs {
			pkgResult, err := ProcessRFPPackage(pkg, path, walkCtx)
			if err != nil {
				return err
			}
			*walkCtx.Results = append(*walkCtx.Results, pkgResult)
			parsed, err := graph.PatchPackageParsed(pkg.ID, walkCtx.Ctx, walkCtx.Cfg)
			if err != nil {
				walkCtx.Cfg.Logger.Info("Failed PATCH request for Package", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division, "error", err)
				return err
			}
			if !parsed.Parsed {
				walkCtx.Cfg.Logger.Info("Failed PATCH request for Package", "Package Name", pkg.Name, "Year", path.Year, "Business Unit", path.BusinessUnit, "Division", path.Division)
				return fmt.Errorf("unable to set Parsed = true: pkg ID: %s, pkg Name: %s", pkg.ID, pkg.Name)
			}
		}
	}

	return nil
}

func removeProcessedPackages(pkgs []graph.Package) ([]graph.Package, error) {
	unprocessedPkgs := make([]graph.Package, 0)
	for _, pkg := range pkgs {
		if pkg.ListItem.Fields.Parsed == false {
			unprocessedPkgs = append(unprocessedPkgs, pkg)
		}
	}

	return unprocessedPkgs, nil
}

func isValidYear(s string) bool {
	year, err := strconv.Atoi(s)
	if err != nil {
		return false
	}

	currentYear := time.Now().Year()
	return year >= 2025 && year <= currentYear+1
}
