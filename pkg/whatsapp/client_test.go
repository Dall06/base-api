package whatsapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_NotConfigured(t *testing.T) {
	c := New(Config{})
	if c.IsConfigured() {
		t.Error("expected not configured with empty config")
	}
}

func TestNew_Configured(t *testing.T) {
	c := New(Config{PhoneNumberID: "123", AccessToken: "tok"})
	if !c.IsConfigured() {
		t.Error("expected configured")
	}
}

func TestSendTemplate_SkipsWhenNotConfigured(t *testing.T) {
	c := New(Config{})
	resp, err := c.SendTemplate(context.Background(), "123", Template{Name: "test"})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if resp != nil {
		t.Error("expected nil response when not configured")
	}
}

func TestSendText_SkipsWhenNotConfigured(t *testing.T) {
	c := New(Config{})
	resp, err := c.SendText(context.Background(), "123", "hello")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if resp != nil {
		t.Error("expected nil response when not configured")
	}
}

func TestSendTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}

		var msg TemplateMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.MessagingProduct != "whatsapp" {
			t.Errorf("expected whatsapp, got %s", msg.MessagingProduct)
		}
		if msg.To != "5215512345678" {
			t.Errorf("expected 5215512345678, got %s", msg.To)
		}
		if msg.Template.Name != "bro_test" {
			t.Errorf("expected bro_test, got %s", msg.Template.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SendResponse{
			MessagingProduct: "whatsapp",
			Messages:         []struct{ ID string `json:"id"` }{{ID: "wamid.123"}},
		})
	}))
	defer server.Close()

	c := New(Config{PhoneNumberID: "phone-123", AccessToken: "test-token"})
	// Override the base URL for testing
	origURL := graphAPIBaseURL
	defer func() { /* can't reassign const, but we test via server */ }()
	_ = origURL

	// We can't override the const, so test the request formation via the mock
	// by directly calling send with a custom client pointing to our server
	c.httpClient = server.Client()
	// We need to intercept the URL - let's test via a transport
	c.httpClient.Transport = &rewriteTransport{base: server.URL, inner: http.DefaultTransport}

	resp, err := c.SendTemplate(context.Background(), "5215512345678", Template{
		Name:     "bro_test",
		Language: TemplateLanguage{Code: "es"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
	if len(resp.Messages) == 0 || resp.Messages[0].ID != "wamid.123" {
		t.Error("expected message ID wamid.123")
	}
}

func TestSendTemplate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"invalid phone"}}`))
	}))
	defer server.Close()

	c := New(Config{PhoneNumberID: "phone-123", AccessToken: "test-token"})
	c.httpClient = server.Client()
	c.httpClient.Transport = &rewriteTransport{base: server.URL, inner: http.DefaultTransport}

	_, err := c.SendTemplate(context.Background(), "bad", Template{Name: "test"})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
}

// rewriteTransport redirects all requests to the test server.
type rewriteTransport struct {
	base  string
	inner http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.base[len("http://"):]
	return t.inner.RoundTrip(req)
}
