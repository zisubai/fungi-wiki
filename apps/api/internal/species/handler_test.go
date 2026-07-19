package species

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type handlerFakeRepo struct {
	item   Species
	params ListParams
}

func (r *handlerFakeRepo) List(_ context.Context, params ListParams) ([]Species, error) {
	r.params = params
	return []Species{r.item}, nil
}
func (r *handlerFakeRepo) Get(context.Context, string) (Species, error)         { return r.item, nil }
func (r *handlerFakeRepo) Create(context.Context, CreateInput) (Species, error) { return r.item, nil }
func (r *handlerFakeRepo) Update(context.Context, string, UpdateInput) (Species, error) {
	return r.item, nil
}
func (r *handlerFakeRepo) Archive(context.Context, string) error { return nil }
func (r *handlerFakeRepo) Delete(context.Context, string) error  { return nil }

func TestPublicDetailHidesDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/species"), &handlerFakeRepo{item: Species{Status: StatusDraft}})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/species/demo", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestPendingReviewCannotBeEdited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAdminRoutes(router.Group("/api/admin/species"), &handlerFakeRepo{item: Species{Status: StatusPendingReview}})
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/species/demo", strings.NewReader(`{"slug":"demo","latinName":"Demo"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, req)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
}

func TestListParsesCombinedFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &handlerFakeRepo{item: Species{Status: StatusPublished}}
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/species"), repo)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/species?functionTag=biocontrol&temperature=30&ph=7&safetyLevel=BSL-1&sourceEnvironment=soil", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if repo.params.FunctionTag != "biocontrol" || repo.params.Temperature == nil || *repo.params.Temperature != 30 || repo.params.PH == nil || *repo.params.PH != 7 || repo.params.SafetyLevel != "BSL-1" || repo.params.SourceEnvironment != "soil" {
		t.Fatalf("unexpected params: %+v", repo.params)
	}
}

func TestListRejectsInvalidPH(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterPublicRoutes(router.Group("/api/species"), &handlerFakeRepo{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/species?ph=15", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
