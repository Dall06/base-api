package publisher

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Publisher defines the interface for event publishing
type Publisher interface {
	Publish(subject string, data interface{}) error
	Close()
}

// NATSPublisher implements Publisher for NATS
type NATSPublisher struct {
	conn *nats.Conn
}

// NewNATSPublisher creates a new NATS publisher
func NewNATSPublisher(url string) (*NATSPublisher, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(nats.DefaultReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	return &NATSPublisher{conn: conn}, nil
}

// Publish sends a message to the specified subject
func (p *NATSPublisher) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := p.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}

	return nil
}

// Close closes the NATS connection
func (p *NATSPublisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}

// NoOpPublisher is a publisher that does nothing (for when NATS is not configured)
type NoOpPublisher struct{}

// NewNoOpPublisher creates a no-op publisher
func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

// Publish does nothing
func (p *NoOpPublisher) Publish(subject string, data interface{}) error {
	return nil
}

// Close does nothing
func (p *NoOpPublisher) Close() {}
