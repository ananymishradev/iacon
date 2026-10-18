package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestGeneratorSendsRequestsAndAppErrorFlag(t *testing.T) {
	var requests atomic.Int64
	var badQuery atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Query().Get("fails") == "app" {
			badQuery.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	gen := &Generator{Client: server.Client(), BaseURL: server.URL, RPS: 20, Mode: "app-errors"}
	done := make(chan struct{})
	go func() { gen.Run(ctx); close(done) }()
	time.Sleep(250 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("generator did not stop after cancellation")
	}
	if requests.Load() == 0 || badQuery.Load() != requests.Load() {
		t.Fatalf("requests=%d app-error requests=%d", requests.Load(), badQuery.Load())
	}
}
