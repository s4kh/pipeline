package lib

import (
	"context"

	"github.com/segmentio/kafka-go"
)

const (
	VOTE_RECEIVED = "vote-received"
	VOTE_GROUP    = "vote-group"
)

type Broker interface {
	ReadMessage(ctx context.Context) ([]byte, error)
}

type Consumer struct {
	*kafka.Reader
}

func NewReader(brokers []string, topic, groupId string) Broker {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupId,
		MinBytes: 10e3,
		MaxBytes: 10e4,
	})

	return &Consumer{Reader: reader}
}

func (r *Consumer) Close() error {
	return r.Reader.Close()
}

func (r *Consumer) ReadMessage(ctx context.Context) ([]byte, error) {
	msg, err := r.Reader.ReadMessage(ctx)

	return msg.Value, err
}
