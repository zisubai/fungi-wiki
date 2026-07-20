package recommendation

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepository struct{ input Input }

func (r *fakeRepository) Recommend(_ context.Context, in Input) (Response, error) {
	r.input = in
	return Response{Items: []Item{}, Disclaimer: "test"}, nil
}
func (r *fakeRepository) RecommendCombination(context.Context, CombinationInput) (CombinationResponse, error) {
	return CombinationResponse{Items: []Combination{}}, nil
}
func (r *fakeRepository) Feedback(context.Context, string, FeedbackInput) error { return nil }
func (r *fakeRepository) CombinationFeedback(context.Context, string, FeedbackInput) error {
	return nil
}
func (r *fakeRepository) CreateCombinationExperiment(context.Context, string, CombinationExperimentInput) (CombinationExperiment, error) {
	return CombinationExperiment{ID: "experiment-1"}, nil
}
func (r *fakeRepository) Quality(context.Context, int) (QualityReport, error) {
	return QualityReport{}, nil
}
func TestRejectsInvalidPH(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/recommendations"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{"requirement":"寻找促生菌","ph":15}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
func TestAcceptsRecommendationRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepository{}
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/recommendations"), repo)
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{"requirement":"寻找土壤生防菌","functionTag":"biocontrol"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 200 || repo.input.FunctionTag != "biocontrol" {
		t.Fatalf("unexpected response %d input %+v", response.Code, repo.input)
	}
}

func TestFeedbackRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterFeedbackRoutes(router.Group("/api/recommendations"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations/not-a-uuid/feedback", strings.NewReader(`{"feedbackType":"helpful"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestCombinationRejectsDuplicateFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/recommendations"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations/combinations", strings.NewReader(`{"functionTags":["biocontrol","biocontrol"]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestCombinationFeedbackRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterFeedbackRoutes(router.Group("/api/recommendations"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations/combinations/not-a-uuid/feedback", strings.NewReader(`{"feedbackType":"helpful"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestCombinationExperimentRejectsInvalidPH(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/api/admin/recommendations"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPost, "/api/admin/recommendations/combinations/550e8400-e29b-41d4-a716-446655440000/experiments", strings.NewReader(`{"candidateIndex":0,"outcome":"compatible","ph":15}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
