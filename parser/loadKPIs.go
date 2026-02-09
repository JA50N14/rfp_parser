package parser

import (
	"os"
	"encoding/json"
	"fmt"
	"regexp"
)

type KPIDefinition struct {
	Name      string           `json:"name"`
	Category  string           `json:"category"`
	Regexps   []*regexp.Regexp `json:"-"`
	RegexStrs []string         `json:"regexps"` //temporary holder
}

const KPIDefPath = "./kpiDefinitions.json"

func LoadKPIDefinitions() ([]KPIDefinition, error) {
	jsonFile, err := os.Open(KPIDefPath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	var kpiDefs []KPIDefinition

	err = decoder.Decode(&kpiDefs)
	if err != nil {
		return nil, err
	}

	if len(kpiDefs) == 0 {
		return nil, fmt.Errorf("kpiDefinition.json file does not contain KPI Definition parsing content")
	}

	return kpiDefs, nil
}

func CompileRegexStrings(kpiDefs []KPIDefinition) ([]KPIDefinition, error) {
	for i := range kpiDefs {
		compiled := make([]*regexp.Regexp, 0, len(kpiDefs[i].RegexStrs))
		for _, pattern := range kpiDefs[i].RegexStrs {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			compiled = append(compiled, re)
		}
		kpiDefs[i].Regexps = compiled
	}
	return kpiDefs, nil
}