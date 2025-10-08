package main

import (
	"regexp"
	"strings"
)

type cleanupRule struct {
	re   *regexp.Regexp
	repl string
}

var cleanupRules = []cleanupRule{
	{regexp.MustCompile(`\n[©•-]\n`), " "},
	{regexp.MustCompile(`[ \t]+`), " "},
	{regexp.MustCompile(`([A-Za-z])\n([a-z])`), "$1$2"},
	{regexp.MustCompile(`([A-Za-z]) \n\n([A-Za-z])`), "$1 $2"},
	{regexp.MustCompile(`([A-Za-z]) \n([A-Za-z])`), "$1 $2"},
	{regexp.MustCompile(`\n \n`), " "},
	{regexp.MustCompile(`\.\n \n`), ". "},
	{regexp.MustCompile(`\n\n+`), "\n"},
	{regexp.MustCompile(`(?m)^[-•]\s*`), ""},
}

var sentenceRule = regexp.MustCompile(`\. [A-Z]`)

func scanTextWithRegex(text string, kpiResults []KpiResult) []KpiResult {
	text = cleanText(text)
	textSlice := strings.Split(text, "\n")

	for _, item := range textSlice {
		for i, kpiResult := range kpiResults {
			for _, re := range kpiResult.Regexps {
				if re.Match([]byte(item)) {
					sentence := extractSentence(item, re, sentenceRule)
					if sentence == "" {
						continue
					}
					kpiResults[i].Sentence = sentence
					kpiResults[i].Found = true
					break
				}
			}
		}
	}
	return kpiResults
}

func cleanText(text string) string {
	for _, rule := range cleanupRules {
		text = rule.re.ReplaceAllString(text, rule.repl)
	}
	return text
}

func extractSentence(text string, target *regexp.Regexp, boundary *regexp.Regexp) string {
	targetLoc := target.FindIndex([]byte(text))
	if targetLoc == nil {
		return ""
	}

	leftBounds := boundary.FindAllStringIndex(text[0:targetLoc[0]], -1)
	leftIdx := 0
	if leftBounds != nil {
		leftBoundsLen := len(leftBounds) - 1
		leftIdx = leftBounds[leftBoundsLen][0] + 2
	}

	rightBound := boundary.FindStringIndex(text[targetLoc[1]:])
	rightIdx := len(text)
	if rightBound != nil {
		rightIdx = targetLoc[1] + rightBound[1] - 2
	}
	return text[leftIdx:rightIdx]
}
