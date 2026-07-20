package audit

import (
	"errors"
	"testing"
)

func TestSubmissionQualityGateReportsMissingRequirements(t *testing.T) {
	err := (submissionQuality{score: 40, safety: true, summary: true}).validate()
	var qualityErr *QualityGateError
	if !errors.As(err, &qualityErr) {
		t.Fatalf("expected quality gate error, got %v", err)
	}
	if qualityErr.Score != 40 || len(qualityErr.Missing) != 4 {
		t.Fatalf("unexpected quality gate details: %+v", qualityErr)
	}
}

func TestSubmissionQualityGateAcceptsCompleteSpecies(t *testing.T) {
	err := (submissionQuality{score: 75, safety: true, summary: true, function: true, culture: true, evidence: true}).validate()
	if err != nil {
		t.Fatalf("expected quality gate to pass: %v", err)
	}
}
