package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"fungi-wiki/apps/api/internal/config"
)

type fakePinger struct{ err error }

func (p fakePinger) Ping(context.Context) error { return p.err }

func TestReadyReportsDatabaseAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name       string
		err        error
		wantCode   int
		wantStatus string
	}{
		{name: "ready", wantCode: http.StatusOK, wantStatus: `"status":"ready"`},
		{name: "unavailable", err: errors.New("database down"), wantCode: http.StatusServiceUnavailable, wantStatus: `"status":"unavailable"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/readyz", Ready(config.Config{AppName: "test-api"}, fakePinger{err: test.err}))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if response.Code != test.wantCode {
				t.Fatalf("status code = %d, want %d", response.Code, test.wantCode)
			}
			if !strings.Contains(response.Body.String(), test.wantStatus) {
				t.Fatalf("response = %s", response.Body.String())
			}
		})
	}
}
