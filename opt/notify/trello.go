// Package notify provides lightweight notification clients for cross-service use.
package notify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// TrelloNotifier sends notifications to Trello as card creation.
type TrelloNotifier struct {
	apiKey     string
	apiToken   string
	listID     string
	httpClient *http.Client
}

// NewTrelloNotifier creates a new Trello notifier.
// If apiKey or apiToken is empty, notifications are silently skipped.
func NewTrelloNotifier(apiKey, apiToken, listID string) *TrelloNotifier {
	return &TrelloNotifier{
		apiKey:     apiKey,
		apiToken:   apiToken,
		listID:     listID,
		httpClient: &http.Client{},
	}
}

// NotifyNewRegistration creates a Trello card for a new gym registration.
func (t *TrelloNotifier) NotifyNewRegistration(ctx context.Context, gymName, slug, ownerEmail, plan, billingCycle string, price float64) error {
	if t.apiKey == "" || t.apiToken == "" || t.listID == "" {
		return nil // skip if not configured (dev mode)
	}

	name := fmt.Sprintf("🏋️ Nuevo registro: %s", gymName)
	desc := fmt.Sprintf(
		"**Gimnasio:** %s\n**Slug:** %s\n**Email:** %s\n**Plan:** %s\n**Ciclo:** %s\n**Precio:** $%.2f MXN",
		gymName, slug, ownerEmail, plan, billingCycle, price,
	)

	params := url.Values{}
	params.Set("key", t.apiKey)
	params.Set("token", t.apiToken)
	params.Set("idList", t.listID)
	params.Set("name", name)
	params.Set("desc", desc)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.trello.com/1/cards?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("trello request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("trello API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("trello error: status %d", resp.StatusCode)
	}

	return nil
}

