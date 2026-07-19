package speciesalias

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepository struct{}

func (*fakeRepository) List(context.Context, string) ([]Alias, error) { return []Alias{}, nil }
func (*fakeRepository) Replace(context.Context, string, []Input) ([]Alias, error) {
	return []Alias{}, nil
}
func TestReplaceRejectsDuplicateAliases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/api/admin/species"), &fakeRepository{})
	request := httptest.NewRequest(http.MethodPut, "/api/admin/species/demo/aliases", strings.NewReader(`{"items":[{"name":"旧名"},{"name":"旧名"}]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
