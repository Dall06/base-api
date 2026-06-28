package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// rewritingTransport redirects any outbound request to target while preserving
// path and query, letting tests intercept calls to api.trello.com.
type rewritingTransport struct {
	target string
}

func (rt rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(rt.target, "http://")
	return http.DefaultTransport.RoundTrip(req)
}

func TestNotifyNewRegistration_SkipsWhenCredentialsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		apiToken string
		listID   string
	}{
		{"all empty", "", "", ""},
		{"only key", "key", "", "list"},
		{"only token", "", "token", "list"},
		{"no list", "key", "token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier := NewTrelloNotifier(tt.apiKey, tt.apiToken, tt.listID)

			err := notifier.NotifyNewRegistration(
				context.Background(),
				"Test Gym", "test-gym", "owner@test.com", "Pro", "monthly", 599.0,
			)
			if err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
		})
	}
}

func TestNotifyNewRegistration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/1/cards" {
			t.Errorf("expected path /1/cards, got %s", r.URL.Path)
		}

		q := r.URL.Query()
		if q.Get("key") != "test-key" {
			t.Errorf("expected key test-key, got %q", q.Get("key"))
		}
		if q.Get("token") != "test-token" {
			t.Errorf("expected token test-token, got %q", q.Get("token"))
		}
		if q.Get("idList") != "list-123" {
			t.Errorf("expected idList list-123, got %q", q.Get("idList"))
		}
		if !strings.Contains(q.Get("name"), "Mi Gym") {
			t.Errorf("expected card name to contain gym name, got %q", q.Get("name"))
		}
		if !strings.Contains(q.Get("desc"), "owner@gym.com") {
			t.Errorf("expected card desc to contain owner email, got %q", q.Get("desc"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewTrelloNotifier("test-key", "test-token", "list-123")
	notifier.httpClient = &http.Client{
		Transport: rewritingTransport{target: server.URL},
	}

	err := notifier.NotifyNewRegistration(
		context.Background(),
		"Mi Gym", "mi-gym", "owner@gym.com", "Pro", "yearly", 5999.0,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyNewRegistration_4xxReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	notifier := NewTrelloNotifier("k", "tok", "list-id")
	notifier.httpClient = &http.Client{
		Transport: rewritingTransport{target: server.URL},
	}

	err := notifier.NotifyNewRegistration(
		context.Background(),
		"Gym", "gym", "a@b.com", "Pro", "monthly", 599.0,
	)
	if err == nil {
		t.Fatal("expected error for 4xx response, got nil")
	}
	if !strings.Contains(err.Error(), "status 401") {
		t.Errorf("expected status 401 in error, got: %v", err)
	}
}

func TestNotifyNewRegistration_CardDescriptionFormat(t *testing.T) {
	var capturedDesc string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedDesc = r.URL.Query().Get("desc")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewTrelloNotifier("k", "tok", "list-id")
	notifier.httpClient = &http.Client{
		Transport: rewritingTransport{target: server.URL},
	}

	_ = notifier.NotifyNewRegistration(
		context.Background(),
		"Power Gym", "power-gym", "admin@power.com", "Pro", "yearly", 5999.0,
	)

	checks := []string{
		"**Gimnasio:** Power Gym",
		"**Slug:** power-gym",
		"**Email:** admin@power.com",
		"**Plan:** Pro",
		"**Ciclo:** yearly",
		"$5999.00 MXN",
	}
	for _, want := range checks {
		if !strings.Contains(capturedDesc, want) {
			t.Errorf("description missing %q, got: %s", want, capturedDesc)
		}
	}
}

func TestNewTrelloNotifier(t *testing.T) {
	n := NewTrelloNotifier("key", "token", "list")
	if n.apiKey != "key" {
		t.Errorf("expected apiKey 'key', got %q", n.apiKey)
	}
	if n.apiToken != "token" {
		t.Errorf("expected apiToken 'token', got %q", n.apiToken)
	}
	if n.listID != "list" {
		t.Errorf("expected listID 'list', got %q", n.listID)
	}
	if n.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
}
