package recommendation

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	temperaturePattern = regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s*(?:°\s*[cC]|[cC]\b|℃)`)
	phPattern          = regexp.MustCompile(`(?i)\bpH\s*[:=为约]?\s*(\d+(?:\.\d+)?)`)
	safetyPattern      = regexp.MustCompile(`(?i)\bBSL\s*[- ]?\s*([1-4])\b`)
)

var sourceKeywords = []string{"土壤", "海洋", "淡水", "食品", "肠道", "植物", "污水", "soil", "marine", "freshwater", "food", "gut"}

type parsedRequirement struct {
	Temperature       *float64
	PH                *float64
	SafetyLevel       string
	SourceEnvironment string
}

func parseRequirement(text string) parsedRequirement {
	result := parsedRequirement{}
	if matches := temperaturePattern.FindStringSubmatch(text); len(matches) > 1 {
		if value, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.Temperature = &value
		}
	}
	if matches := phPattern.FindStringSubmatch(text); len(matches) > 1 {
		if value, err := strconv.ParseFloat(matches[1], 64); err == nil && value >= 0 && value <= 14 {
			result.PH = &value
		}
	}
	if matches := safetyPattern.FindStringSubmatch(text); len(matches) > 1 {
		result.SafetyLevel = "BSL-" + matches[1]
	}
	lowerText := strings.ToLower(text)
	for _, keyword := range sourceKeywords {
		if strings.Contains(lowerText, strings.ToLower(keyword)) {
			result.SourceEnvironment = keyword
			break
		}
	}
	return result
}

func mergeParsedRequirement(input Input, parsed parsedRequirement) Input {
	if input.Temperature == nil {
		input.Temperature = parsed.Temperature
	}
	if input.PH == nil {
		input.PH = parsed.PH
	}
	if strings.TrimSpace(input.SafetyLevel) == "" {
		input.SafetyLevel = parsed.SafetyLevel
	}
	if strings.TrimSpace(input.SourceEnvironment) == "" {
		input.SourceEnvironment = parsed.SourceEnvironment
	}
	return input
}
