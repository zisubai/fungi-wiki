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
