package narrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type LLM struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
	Client   *http.Client
}

func (l LLM) Narrate(ctx context.Context, input Input) (Output, error) {
	payload := map[string]any{
		"model":  l.Model,
		"input":  Redact(input.Bundle),
		"schema": map[string]any{"type": "object", "required": []string{"summary", "cause", "citedEvidence", "confidence", "nextChecks"}},
	}
	if input.RuleVerdict != nil {
		payload["rule_verdict"] = input.RuleVerdict
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return Output{}, err
	}
	base := l.BaseURL
	if base == "" {
		return Output{}, fmt.Errorf("narrator base URL is required")
	}
	client := l.Client
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/"), bytes.NewReader(data))
	if err != nil {
		return Output{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if l.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+l.APIKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Output{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Output{}, fmt.Errorf("narrator status %d", resp.StatusCode)
	}
	var out Output
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Output{}, err
	}
	return out, nil
}
