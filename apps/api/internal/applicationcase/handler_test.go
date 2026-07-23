package applicationcase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeRepository struct{ items []Case }

func (f *fakeRepository) List(context.Context, string) ([]Case, error) { return f.items, nil }
func (f *fakeRepository) Create(_ context.Context, speciesID string, input Input) (Case, error) {
	item := Case{ID: "case-1", SpeciesID: speciesID, Industry: input.Industry, Scenario: input.Scenario}
	f.items = append(f.items, item)
	return item, nil
}
func (f *fakeRepository) Update(context.Context, string, string, Input) (Case, error) {
	return Case{}, nil
}
func (f *fakeRepository) Delete(context.Context, string, string) error { return nil }

func TestCreateAndListApplicationCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &fakeRepository{}
	router := gin.New()
	RegisterAdminRoutes(router.Group("/species"), repository)
	request := httptest.NewRequest(http.MethodPost, "/species/bacillus/application-cases", strings.NewReader(`{"industry":"农业","scenario":"土传病害防控"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	var created Case
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Scenario != "土传病害防控" {
		t.Fatalf("unexpected scenario %q", created.Scenario)
	}
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, httptest.NewRequest(http.MethodGet, "/species/bacillus/application-cases", nil))
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listResponse.Code)
	}
}

func TestCreateApplicationCaseValidatesRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/species"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPost, "/species/bacillus/application-cases", strings.NewReader(`{"industry":"农业"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
