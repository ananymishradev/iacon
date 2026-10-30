package httpx

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

const Version = "iacon-ship1"

type Server struct {
	mu       sync.RWMutex
	now      func() time.Time
	lastPoll *time.Time
	incident json.RawMessage
	stopping bool
}

func NewServer() *Server {
	return &Server{now: time.Now}
}

func NewServerWithClock(now func() time.Time) *Server {
	if now == nil {
		now = time.Now
	}
	return &Server{now: now}
}

func (s *Server) SetTick(unixMs int64) {
	t := time.UnixMilli(unixMs)
	s.mu.Lock()
	s.lastPoll = &t
	s.mu.Unlock()
}

func (s *Server) SetIncident(snapshot []byte) {
	s.mu.Lock()
	if len(snapshot) == 0 || string(snapshot) == "null" {
		s.incident = nil
	} else {
		s.incident = append(json.RawMessage(nil), snapshot...)
	}
	s.mu.Unlock()
}

func (s *Server) SetStopping() {
	s.mu.Lock()
	s.stopping = true
	s.mu.Unlock()
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/live", s.live)
	mux.HandleFunc("/ready", s.ready)
	mux.HandleFunc("/health", s.ready)
	return mux
}

func (s *Server) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "alive", "version": Version})
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	lastPoll := s.lastPoll
	stopping := s.stopping
	incident := append(json.RawMessage(nil), s.incident...)
	now := s.now()
	s.mu.RUnlock()
	ready := !stopping && lastPoll != nil && now.Sub(*lastPoll) <= 45*time.Second && now.Sub(*lastPoll) >= 0
	body := struct {
		Version  string          `json:"version"`
		Ready    bool            `json:"ready"`
		LastPoll *int64          `json:"lastPoll"`
		Incident json.RawMessage `json:"incident"`
	}{Version: Version, Ready: ready, Incident: incident}
	if lastPoll != nil {
		ms := lastPoll.UnixMilli()
		body.LastPoll = &ms
	}
	if len(body.Incident) == 0 {
		body.Incident = json.RawMessage("null")
	}
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, body)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
