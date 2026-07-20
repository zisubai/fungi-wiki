package recommendation

import (
	"strings"
	"testing"
)

func TestCandidateScoreIncludesEvidenceAndRisk(t *testing.T) {
	candidate := candidate{item: Item{SafetyLevel: "BSL-2", EvidenceCount: 3}, quality: 80, evidenceAverage: 70, functionConfidence: 90, functionName: "生防"}
	candidate.score(Input{}, "biocontrol")
	if candidate.item.Score <= 0 {
		t.Fatal("expected positive score")
	}
	if candidate.item.RiskWarning == "" {
		t.Fatal("expected risk warning")
	}
	if len(candidate.item.Reasons) < 3 {
		t.Fatalf("expected explainable reasons: %+v", candidate.item.Reasons)
	}
	if !strings.Contains(candidate.item.Reasons[0], "生防") {
		t.Fatalf("missing function reason: %+v", candidate.item.Reasons)
	}
}

func TestBuildRelaxationSuggestionsIdentifiesBlockingConditions(t *testing.T) {
	temperature, ph := 30.0, 7.0
	input := Input{Temperature: &temperature, PH: &ph, SourceEnvironment: "土壤"}
	suggestions := buildRelaxationSuggestions(input, "biocontrol", diagnosticCounts{total: 10, safety: 10})
	if len(suggestions) != 4 {
		t.Fatalf("expected four blocking suggestions, got %+v", suggestions)
	}
	if !strings.Contains(strings.Join(suggestions, " "), "功能") || !strings.Contains(strings.Join(suggestions, " "), "温度") {
		t.Fatalf("unexpected suggestions: %+v", suggestions)
	}
}

func TestBuildCombinationCalculatesCultureOverlap(t *testing.T) {
	firstMin, firstMax, secondMin, secondMax := 25.0, 37.0, 20.0, 30.0
	firstPHMin, firstPHMax, secondPHMin, secondPHMax := 6.0, 8.0, 4.0, 6.0
	combination := buildCombination(
		combinationCandidate{member: CombinationMember{ID: "one", EvidenceCount: 1}, quality: 90, temperatureMin: &firstMin, temperatureMax: &firstMax, phMin: &firstPHMin, phMax: &firstPHMax},
		combinationCandidate{member: CombinationMember{ID: "two", EvidenceCount: 1}, quality: 80, temperatureMin: &secondMin, temperatureMax: &secondMax, phMin: &secondPHMin, phMax: &secondPHMax},
	)
	if !combination.Compatible || *combination.TemperatureMin != 25 || *combination.TemperatureMax != 30 || *combination.PHMin != 6 || *combination.PHMax != 6 {
		t.Fatalf("unexpected combination overlap: %+v", combination)
	}
}

func TestRankCombinationsPrefersCompatiblePair(t *testing.T) {
	min20, max24, max30, min40, max45, ph6, ph8 := 20.0, 24.0, 30.0, 40.0, 45.0, 6.0, 8.0
	first := []combinationCandidate{
		{member: CombinationMember{ID: "high"}, quality: 100, temperatureMin: &min40, temperatureMax: &max45, phMin: &ph6, phMax: &ph8},
		{member: CombinationMember{ID: "compatible"}, quality: 80, temperatureMin: &min20, temperatureMax: &max24, phMin: &ph6, phMax: &ph8},
	}
	second := []combinationCandidate{{member: CombinationMember{ID: "target"}, quality: 90, temperatureMin: &min20, temperatureMax: &max30, phMin: &ph6, phMax: &ph8}}
	items := rankCombinations(first, second, 5)
	if len(items) != 2 || !items[0].Compatible || items[0].Members[0].ID != "compatible" {
		t.Fatalf("expected compatible pair first: %+v", items)
	}
}
