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

	app := New(ctx, log, 18889, 44044, false, time.Hour)

	go func() {
		app.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:18889/metrics")
	if err == nil {
		resp.Body.Close()
	}

	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = app.Stop(stopCtx)
	if err != nil {
		t.Errorf("expected graceful stop, got error: %v", err)
	}

	_, err = http.Get("http://localhost:18889/metrics")
	if err == nil {
		t.Error("expected server to be down, but request succeeded")
	}
}
