package httpserver

import (
	"testing"

	"fungi-wiki/apps/api/internal/config"
)

func TestSplitCommaSeparated(t *testing.T) {
	items := splitCommaSeparated("127.0.0.1, ::1, 10.0.0.0/8, ")
	want := []string{"127.0.0.1", "::1", "10.0.0.0/8"}
	if len(items) != len(want) {
		t.Fatalf("items = %#v", items)
	}
	for index := range want {
		if items[index] != want[index] {
			t.Fatalf("items[%d] = %q, want %q", index, items[index], want[index])
		}
	}
}

func TestNewRouterRejectsInvalidTrustedProxy(t *testing.T) {
	_, err := NewRouter(config.Config{TrustedProxies: "not-an-ip"}, nil)
	if err == nil {
		t.Fatal("expected invalid trusted proxy error")
	}
}
