package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type PrometheusCollector struct {
	URL    string
	Client *http.Client
}

func (c PrometheusCollector) Name() string { return "prometheus" }

func (c PrometheusCollector) Collect(ctx context.Context) (map[string]any, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	data := map[string]any{"prom_ok": true}
	for key, query := range map[string]string{
		"req_30s":       "sum(increase(workshop_http_requests_total[30s]))",
		"err_ratio_30s": `sum(increase(workshop_http_requests_total{status=~"5.."}[30s])) / sum(increase(workshop_http_requests_total[30s]))`,
	} {
		value, ok, err := promQuery(ctx, client, c.URL, query)
		if err != nil {
			data["prom_ok"], data["unknown_data"] = false, true
			return data, nil
		}
		if !ok {
			data["unknown_data"] = true
			continue
		}
		data[key] = value
	}
	return data, nil
}

func promQuery(ctx context.Context, client *http.Client, baseURL, query string) (float64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimRight(baseURL, "/")+"/api/v1/query?query="+urlQueryEscape(query), nil)
	if err != nil {
		return 0, false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, false, fmt.Errorf("prometheus status %d", resp.StatusCode)
	}
	var body struct {
		Data struct {
			Result []struct {
				Value []any `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, false, err
	}
	if len(body.Data.Result) == 0 || len(body.Data.Result[0].Value) < 2 {
		return 0, false, nil
	}
	value, ok := body.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, false, nil
	}
	var parsed float64
	_, err = fmt.Sscan(value, &parsed)
	return parsed, err == nil, err
}

func urlQueryEscape(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, " ", "+"), `"`, "%22")
}

type ClickHouseCollector struct {
	URL      string
	User     string
	Password string
	Client   *http.Client
}

func (c ClickHouseCollector) Name() string { return "clickhouse" }

func (c ClickHouseCollector) Collect(ctx context.Context) (map[string]any, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	query := "SELECT ServiceName, SpanName, StatusCode, Attributes['workshop.failures'] FROM otel.otel_traces WHERE Timestamp > now() - INTERVAL 5 MINUTE AND StatusCode = 'Error' LIMIT 20 FORMAT JSON"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.User, c.Password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("clickhouse status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return map[string]any{"ch_ok": true, "error_spans": payload.Data}, nil
}
