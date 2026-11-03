package signals

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HTTPProbe(baseURL, path string, timeout time.Duration) Sample {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+path, nil)
	if err != nil {
		return Sample{At: time.Now(), State: StateBAD, Definitive: false, Detail: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		definitive := !errorsIsTimeout(err)
		return Sample{At: time.Now(), State: StateBAD, Definitive: definitive, Detail: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	state := StateOK
	if resp.StatusCode >= 500 {
		state = StateBAD
	}
	return Sample{At: time.Now(), State: state, Detail: string(body)}
}

func FastAppReadySample(baseURL string) Sample { return HTTPProbe(baseURL, "/ready", 2*time.Second) }
func ReadSample(baseURL string) Sample         { return HTTPProbe(baseURL, "/get", 2*time.Second) }

func errorsIsTimeout(err error) bool {
	if timeout, ok := err.(net.Error); ok {
		return timeout.Timeout()
	}
	return strings.Contains(strings.ToLower(err.Error()), "timeout")
}

type CanaryStore interface {
	Ensure(context.Context) error
	Write(context.Context, int64) error
}

func CanarySample(ctx context.Context, store CanaryStore, seq int64) Sample {
	now := time.Now()
	if err := store.Ensure(ctx); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			return Sample{At: now, State: StateUNKNOWN, Detail: "canary bootstrap failed: permission"}
		}
		return classifyCanaryError(now, err)
	}
	if err := store.Write(ctx, seq); err != nil {
		return classifyCanaryError(now, err)
	}
	return Sample{At: now, State: StateOK}
}

func classifyCanaryError(now time.Time, err error) Sample {
	text := strings.ToLower(err.Error())
	definitive := strings.Contains(text, "53100") || strings.Contains(text, "no space left") ||
		strings.Contains(text, "connection refused") || strings.Contains(text, "sqlstate 08") ||
		strings.Contains(text, "sqlstate 57")
	return Sample{At: now, State: StateBAD, Definitive: definitive, Detail: err.Error()}
}

type PromClient struct {
	URL    string
	Client *http.Client
}

func (p PromClient) Sample(ctx context.Context) Sample {
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(p.URL, "/")+"/api/v1/query", nil)
	if err != nil {
		return Sample{At: time.Now(), State: StateUNKNOWN, Detail: "prometheus unreachable"}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Sample{At: time.Now(), State: StateUNKNOWN, Detail: "prometheus unreachable"}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Sample{At: time.Now(), State: StateUNKNOWN, Detail: "prometheus unreachable"}
	}
	var payload struct {
		Data struct {
			Result []struct {
				Value []json.RawMessage `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || len(payload.Data.Result) == 0 {
		return Sample{At: time.Now(), State: StateUNKNOWN, Detail: "prometheus unreachable"}
	}
	return Sample{At: time.Now(), State: StateOK}
}

func ParsePromRatio(req, failures float64) Sample {
	if req < 20 {
		return Sample{At: time.Now(), State: StateUNKNOWN, Detail: "low traffic"}
	}
	if req > 0 && failures/req > 0.02 {
		return Sample{At: time.Now(), State: StateBAD, Detail: fmt.Sprintf("error ratio %.4f", failures/req)}
	}
	return Sample{At: time.Now(), State: StateOK, Detail: strconv.FormatFloat(failures/req, 'f', 4, 64)}
}
