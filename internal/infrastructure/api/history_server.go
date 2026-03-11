package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	historydomain "game-time-tracker/internal/domain/history"
)

// HistoryStore is the server persistence port.
type HistoryStore interface {
	Load() ([]historydomain.Entry, error)
	Save(entries []historydomain.Entry) error
}

type HistoryServer struct {
	apiKey string
	store  HistoryStore
	mu     sync.Mutex
}

type historyPayload struct {
	Entries []historydomain.Entry `json:"entries"`
}

func NewHistoryServer(apiKey string, store HistoryStore) *HistoryServer {
	return &HistoryServer{apiKey: strings.TrimSpace(apiKey), store: store}
}

func (s *HistoryServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/history", s.handleHistory)
	mux.HandleFunc("/v1/history/append", s.handleAppend)
	return mux
}

func (s *HistoryServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *HistoryServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	s.mu.Lock()
	entries, err := s.store.Load()
	s.mu.Unlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("load history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(historyPayload{Entries: entries}); err != nil {
		http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *HistoryServer) handleAppend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	var payload historyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}
	if len(payload.Entries) == 0 {
		http.Error(w, "entries must not be empty", http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entries {
		if strings.TrimSpace(entry.GameName) == "" {
			http.Error(w, "entry gameName must not be empty", http.StatusBadRequest)
			return
		}
		if entry.TotalTimeSecs <= 0 {
			http.Error(w, "entry totalTimeSecs must be greater than zero", http.StatusBadRequest)
			return
		}
	}

	s.mu.Lock()
	err := s.store.Save(payload.Entries)
	s.mu.Unlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("save entries: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HistoryServer) authorized(r *http.Request) bool {
	if s.apiKey == "" {
		return true
	}
	token, err := bearerToken(r)
	if err == nil {
		return token == s.apiKey
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key")) == s.apiKey
}

func bearerToken(r *http.Request) (string, error) {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if value == "" {
		return "", errors.New("missing authorization header")
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}
	if strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("empty bearer token")
	}
	return strings.TrimSpace(parts[1]), nil
}
