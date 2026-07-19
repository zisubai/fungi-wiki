package culturecondition

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepo struct{}

func (*fakeRepo) List(context.Context, string) ([]Condition, error) { return []Condition{}, nil }
func (*fakeRepo) Replace(context.Context, string, []Input) ([]Condition, error) {
	return []Condition{}, nil
}
func TestRejectsInvalidPH(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAdminRoutes(r.Group("/api/admin/species"), &fakeRepo{})
	w := httptest.NewRecorder()
	q := httptest.NewRequest(http.MethodPut, "/api/admin/species/demo/culture-conditions", strings.NewReader(`{"items":[{"phMin":15}]}`))
	q.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, q)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
