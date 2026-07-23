package speciesversion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeRepository struct{ limit int }

func (f *fakeRepository) List(_ context.Context, _ string, limit int) ([]Version, error) {
	f.limit = limit
	return []Version{{ID: "v1", VersionNumber: 1, ChangeType: "baseline", SourceTable: "migration"}}, nil
}

func TestListSpeciesVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &fakeRepository{}
	router := gin.New()
	RegisterAdminRoutes(router.Group("/species"), repository)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/species/bacillus/versions?limit=12", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if repository.limit != 12 {
		t.Fatalf("expected limit 12, got %d", repository.limit)
	}
}
