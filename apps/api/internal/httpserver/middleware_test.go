package httpserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name     string
		incoming string
		want     string
	}{
		{name: "preserves valid incoming ID", incoming: "trace-123", want: "trace-123"},
		{name: "replaces invalid incoming ID", incoming: "invalid id\n", want: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(requestIDMiddleware())
			router.GET("/", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("X-Request-ID", test.incoming)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			got := response.Header().Get("X-Request-ID")
			if test.want != "" && got != test.want {
				t.Fatalf("request ID = %q, want %q", got, test.want)
			}
			if test.want == "" && !requestIDPattern.MatchString(got) {
				t.Fatalf("generated invalid request ID %q", got)
			}
		})
	}
}

func TestRequestLoggerIncludesCorrelationFields(t *testing.T) {
	var output bytes.Buffer
	previousWriter := gin.DefaultWriter
	gin.DefaultWriter = &output
	t.Cleanup(func() { gin.DefaultWriter = previousWriter })
	router := gin.New()
	router.Use(requestIDMiddleware(), requestLoggerMiddleware())
	router.GET("/test", func(ctx *gin.Context) { ctx.Status(http.StatusCreated) })
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("X-Request-ID", "trace-logger-1")
	router.ServeHTTP(httptest.NewRecorder(), request)
	logLine := output.String()
	for _, expected := range []string{"request_id=trace-logger-1", "method=GET", "path=/test", "status=201", "client_ip="} {
		if !bytes.Contains([]byte(logLine), []byte(expected)) {
			t.Errorf("log %q does not contain %q", logLine, expected)
		}
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(securityHeadersMiddleware())
	router.GET("/", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	for header, want := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Referrer-Policy": "no-referrer", "Permissions-Policy": "camera=(), microphone=(), geolocation=()"} {
		if got := response.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}
