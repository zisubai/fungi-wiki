package speciesfunction

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeRepository struct {
	items []SpeciesFunction
	err   error
}

func (repo *fakeRepository) List(context.Context, string) ([]SpeciesFunction, error) {
	return repo.items, repo.err
}

func (repo *fakeRepository) Replace(_ context.Context, _ string, items []ReplaceItem) ([]SpeciesFunction, error) {
	if repo.err != nil {
		return nil, repo.err
	}
	result := make([]SpeciesFunction, len(items))
	for index, item := range items {
		result[index] = SpeciesFunction{FunctionTagID: item.FunctionTagID, ConfidenceScore: item.ConfidenceScore}
	}
	return result, nil
}

func TestReplaceRejectsInvalidConfidenceScore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/api/admin/species"), &fakeRepository{})

	request := httptest.NewRequest(http.MethodPut, "/api/admin/species/demo/functions",
		strings.NewReader(`{"items":[{"functionTagId":"tag-id","confidenceScore":101}]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestListReturnsItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/species"), &fakeRepository{items: []SpeciesFunction{{FunctionTagID: "tag-id", FunctionTagName: "促生"}}})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/species/demo/functions", nil))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "促生") {
		t.Fatalf("unexpected response %d: %s", response.Code, response.Body.String())
	}
}
