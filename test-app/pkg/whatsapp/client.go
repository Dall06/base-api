// Package whatsapp provides a lightweight client for the Meta Cloud API
// (WhatsApp Business Platform). It handles sending template and text
// messages through the Graph API without any third-party BSP middleman.
package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const graphAPIBaseURL = "https://graph.facebook.com/v21.0"

// Config holds the credentials for Meta Cloud API.
type Config struct {
	PhoneNumberID string // The WhatsApp phone-number ID (not the phone number itself)
	AccessToken   string // Permanent system-user token from Meta Business Manager
}

// Client is a thin wrapper around the Meta Graph API for WhatsApp messaging.
type Client struct {
	cfg        Config
	httpClient *http.Client
}

// New creates a new WhatsApp client. If PhoneNumberID or AccessToken is empty
// the client is created but all Send* methods will silently return nil (dev mode).
func New(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true when credentials are present.
func (c *Client) IsConfigured() bool {
	return c.cfg.PhoneNumberID != "" && c.cfg.AccessToken != ""
}

// ---- request / response types ------------------------------------------------

// TemplateMessage represents a WhatsApp template message payload.
type TemplateMessage struct {
	MessagingProduct string   `json:"messaging_product"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Template         Template `json:"template"`
}

// Template is the template section of a message.
type Template struct {
	Name       string              `json:"name"`
	Language   TemplateLanguage    `json:"language"`
	Components []TemplateComponent `json:"components,omitempty"`
}

// TemplateLanguage specifies the language code (e.g. "es", "en").
type TemplateLanguage struct {
	Code string `json:"code"`
}

// TemplateComponent is a header/body/button component with parameters.
type TemplateComponent struct {
	Type       string              `json:"type"`
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

// TemplateParameter is a single parameter value inside a component.
type TemplateParameter struct {
	Type  string         `json:"type"`
	Text  string         `json:"text,omitempty"`
	Image *MediaObject   `json:"image,omitempty"`
}

// MediaObject references a media asset (image, document, etc.).
type MediaObject struct {
	Link string `json:"link,omitempty"`
}

// TextMessage represents a plain text message payload.
type TextMessage struct {
	MessagingProduct string      `json:"messaging_product"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             TextContent `json:"text"`
}

// TextContent is the text body of a message.
type TextContent struct {
	Body       string `json:"body"`
	PreviewURL bool   `json:"preview_url,omitempty"`
}

// SendResponse is the API response after sending a message.
type SendResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

// APIError is returned when the Graph API responds with an error.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("whatsapp api error (status %d): %s", e.StatusCode, e.Body)
}

// ---- public send methods ----------------------------------------------------

// SendTemplate sends a pre-approved template message.
func (c *Client) SendTemplate(ctx context.Context, to string, tmpl Template) (*SendResponse, error) {
	if !c.IsConfigured() {
		return nil, nil
	}

	msg := TemplateMessage{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "template",
		Template:         tmpl,
	}

	return c.send(ctx, msg)
}

// SendText sends a free-form text message (only within the 24-hour window).
func (c *Client) SendText(ctx context.Context, to, body string) (*SendResponse, error) {
	if !c.IsConfigured() {
		return nil, nil
	}

	msg := TextMessage{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
		Text:             TextContent{Body: body},
	}

	return c.send(ctx, msg)
}

// ---- internal ---------------------------------------------------------------

func (c *Client) send(ctx context.Context, payload any) (*SendResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal whatsapp payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", graphAPIBaseURL, c.cfg.PhoneNumberID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create whatsapp request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("whatsapp api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read whatsapp response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var result SendResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal whatsapp response: %w", err)
	}

	return &result, nil
}
