package evidence

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type AppCollector struct {
	URL    string
	Client *http.Client
}

func (c AppCollector) Name() string { return "app" }

func (c AppCollector) Collect(ctx context.Context) (map[string]any, error) {
	client := c.Client
	if client == nil {
		client = &http.Client{}
	}
	data := map[string]any{}
	for _, path := range []string{"/ready", "/get"} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.URL, "/")+path, nil)
		if err != nil {
			data[pathKey(path)+"_err"] = err.Error()
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			data[pathKey(path)+"_err"] = err.Error()
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		resp.Body.Close()
		if path == "/ready" {
			data["ready_status"], data["ready_body"] = resp.StatusCode, string(body)
		} else {
			data["get_status"] = resp.StatusCode
			if resp.StatusCode == http.StatusOK {
				data["item_count"] = strings.TrimSpace(string(body))
			}
		}
	}
	return data, nil
}

func pathKey(path string) string {
	if path == "/ready" {
		return "ready"
	}
	return "get"
}

type RedisCollector struct {
	Addr    string
	Timeout time.Duration
}

func (c RedisCollector) Name() string { return "redis" }

func (c RedisCollector) Collect(ctx context.Context) (map[string]any, error) {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 1500 * time.Millisecond
	}
	dialer := netDialer{timeout: timeout}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", c.Addr)
	if err != nil {
		return map[string]any{"ping_ok": false, "err": err.Error(), "latency_ms": time.Since(start).Milliseconds()}, nil
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return map[string]any{"ping_ok": false, "err": err.Error(), "latency_ms": time.Since(start).Milliseconds()}, nil
	}
	buf := make([]byte, 64)
	if _, err := conn.Read(buf); err != nil {
		return map[string]any{"ping_ok": false, "err": err.Error(), "latency_ms": time.Since(start).Milliseconds()}, nil
	}
	return map[string]any{"ping_ok": true, "latency_ms": time.Since(start).Milliseconds()}, nil
}

type netDialer struct{ timeout time.Duration }

func (d netDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var nd net.Dialer
	nd.Timeout = d.timeout
	return nd.DialContext(ctx, network, address)
}
