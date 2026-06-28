package subscriber

import (
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// Handler defines the interface for message handlers
type Handler interface {
	Handle(msg *nats.Msg) error
}

// NATSSubscriber manages NATS subscriptions
type NATSSubscriber struct {
	conn          *nats.Conn
	subscriptions []*nats.Subscription
}

// NewNATSSubscriber creates a new NATS subscriber
func NewNATSSubscriber(url string) (*NATSSubscriber, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(nats.DefaultReconnectWait),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				slog.Warn("NATS disconnected", slog.String("error", err.Error()))
				return
			}
			slog.Warn("NATS disconnected")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			slog.Info("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, err
	}

	return &NATSSubscriber{
		conn:          conn,
		subscriptions: make([]*nats.Subscription, 0),
	}, nil
}

// Subscribe subscribes to a subject with a queue group
func (s *NATSSubscriber) Subscribe(subject, queue string, handler Handler) error {
	sub, err := s.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		if err := handler.Handle(msg); err != nil {
			slog.Error("message handling failed",
				slog.String("subject", subject),
				slog.String("error", err.Error()),
			)
		}
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	slog.Info("subscribed to subject",
		slog.String("subject", subject),
		slog.String("queue", queue),
	)

	return nil
}

// Publish publishes a message to a subject
func (s *NATSSubscriber) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Publish(subject, payload)
}

// Close closes all subscriptions and the connection
func (s *NATSSubscriber) Close() {
	for _, sub := range s.subscriptions {
		_ = sub.Unsubscribe()
	}
	s.conn.Close()
}

// IsConnected returns true if connected to NATS
func (s *NATSSubscriber) IsConnected() bool {
	return s.conn.IsConnected()
}
