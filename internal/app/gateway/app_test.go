package gateway

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestHTTPGateway_GracefulShutdown(t *testing.T) {
	ctx := context.Background()
	log := slog.Default()

	// Use random high port to avoid conflicts
	app := New(ctx, log, 18889, 44044, false, time.Hour)

	go func() {
		app.Run()
	}()

	// Allow server to start
	time.Sleep(100 * time.Millisecond)

	// Send a dummy request to ensure it's up
	resp, err := http.Get("http://localhost:18889/metrics")
	if err == nil {
		resp.Body.Close()
	}

	// Trigger stop
	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = app.Stop(stopCtx)
	if err != nil {
		t.Errorf("expected graceful stop, got error: %v", err)
	}

	// Ensure server is actually down
	_, err = http.Get("http://localhost:18889/metrics")
	if err == nil {
		t.Error("expected server to be down, but request succeeded")
	}
}
