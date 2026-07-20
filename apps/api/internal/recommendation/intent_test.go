package recommendation

import "testing"

func TestParseRequirementExtractsStructuredConditions(t *testing.T) {
	parsed := parseRequirement("寻找适合 30°C、pH 7.2 的土壤生防菌，要求 BSL-1")
	if parsed.Temperature == nil || *parsed.Temperature != 30 || parsed.PH == nil || *parsed.PH != 7.2 || parsed.SafetyLevel != "BSL-1" || parsed.SourceEnvironment != "土壤" {
		t.Fatalf("unexpected parsed requirement: %+v", parsed)
	}
}

func TestExplicitInputOverridesParsedRequirement(t *testing.T) {
	temperature, ph := 25.0, 6.5
	input := mergeParsedRequirement(Input{Temperature: &temperature, PH: &ph, SafetyLevel: "BSL-2", SourceEnvironment: "海洋"}, parseRequirement("30℃ pH 7 土壤 BSL-1"))
	if *input.Temperature != 25 || *input.PH != 6.5 || input.SafetyLevel != "BSL-2" || input.SourceEnvironment != "海洋" {
		t.Fatalf("explicit input was overwritten: %+v", input)
	}
}
