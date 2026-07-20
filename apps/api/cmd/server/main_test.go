package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServerConfiguresTimeouts(t *testing.T) {
	server := newHTTPServer(":0", http.NewServeMux())
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second || server.WriteTimeout != 30*time.Second || server.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected server timeouts: %#v", server)
	}
}

func TestRunHTTPServerStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	server := newHTTPServer("127.0.0.1:0", http.NewServeMux())
	if err := runHTTPServer(ctx, server, time.Second); err != nil {
		t.Fatalf("runHTTPServer returned error: %v", err)
	}
}
