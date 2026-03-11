package history

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	historydomain "game-time-tracker/internal/domain/history"
)

type RemoteRepository struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type remoteHistoryPayload struct {
	Entries []historydomain.Entry `json:"entries"`
}

func NewRemoteRepository(baseURL, apiKey string, client *http.Client) *RemoteRepository {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &RemoteRepository{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		client:  client,
	}
}

func (r *RemoteRepository) Load() ([]historydomain.Entry, error) {
	endpoint := r.baseURL + "/v1/history"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	r.addAuth(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request history: unexpected status %d", resp.StatusCode)
	}

	var payload remoteHistoryPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode history: %w", err)
	}
	return payload.Entries, nil
}

func (r *RemoteRepository) Save(entries []historydomain.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	endpoint := r.baseURL + "/v1/history/append"
	payload := remoteHistoryPayload{Entries: entries}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal entries: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r.addAuth(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("send entries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("send entries: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (r *RemoteRepository) addAuth(req *http.Request) {
	if r.apiKey == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
}
