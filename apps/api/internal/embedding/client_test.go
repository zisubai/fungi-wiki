package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientUsesOpenAICompatibleEmbeddingProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing bearer token")
		}
		var body struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "semantic-model" || len(body.Input) != 2 {
			t.Fatalf("unexpected request: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]any{"index": 0, "embedding": []float32{1, 0}}, map[string]any{"index": 1, "embedding": []float32{0, 1}}}})
	}))
	defer server.Close()
	client := NewClient(server.URL, "secret", "semantic-model")
	vectors, err := client.Embed(context.Background(), []string{"生防", "发酵"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 2 || vectors[0][0] != 1 || vectors[1][1] != 1 {
		t.Fatalf("unexpected vectors: %#v", vectors)
	}
}

func TestClientReportsDisabledWithoutEndpoint(t *testing.T) {
	client := NewClient("", "", "model")
	if client.Enabled() {
		t.Fatal("expected disabled client")
	}
	if _, err := client.Embed(context.Background(), []string{"query"}); err != ErrDisabled {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}
