package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type Generator struct {
	Client   *http.Client
	BaseURL  string
	RPS      int
	Mode     string
	Requests atomic.Int64
	Failures atomic.Int64
}

func (g *Generator) Run(ctx context.Context) {
	if g.Client == nil {
		g.Client = &http.Client{}
	}
	if g.RPS < 1 {
		g.RPS = 1
	}
	ticker := time.NewTicker(time.Second / time.Duration(g.RPS))
	defer ticker.Stop()
	summary := time.NewTicker(10 * time.Second)
	defer summary.Stop()
	var sequence uint64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sequence++
			go g.request(ctx, sequence)
		case <-summary.C:
			log.Printf("loadgen requests=%d failures=%d", g.Requests.Load(), g.Failures.Load())
		}
	}
}

func (g *Generator) request(parent context.Context, sequence uint64) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	path := "/get"
	if sequence%5 == 0 {
		path = "/post"
	}
	if g.Mode == "app-errors" {
		path += "?fails=app"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.BaseURL+path, nil)
	if path == "/post" || path == "/post?fails=app" {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, g.BaseURL+path, nil)
	}
	if err != nil {
		g.Failures.Add(1)
		return
	}
	resp, err := g.Client.Do(req)
	g.Requests.Add(1)
	if err != nil {
		g.Failures.Add(1)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		g.Failures.Add(1)
	}
}

func main() {
	rps, _ := strconv.Atoi(os.Getenv("RPS"))
	if rps == 0 {
		rps = 6
	}
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://app:8080"
	}
	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "normal"
	}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	(&Generator{BaseURL: baseURL, RPS: rps, Mode: mode}).Run(ctx)
}
