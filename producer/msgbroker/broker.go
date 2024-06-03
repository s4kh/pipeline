package msgbroker

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type MsgBroker interface {
	Publish(msg, key, topic string, res chan<- PublishRes)
}

const (
	VOTE_RECEIVED = "vote-received"
)

type KMsgBroker struct {
	Host   string
	Port   string
	Writer kafka.Writer
}

type PublishRes struct {
	code int // 0 - success, 1 - error
	err  error
}

func (pr PublishRes) String() string {
	return fmt.Sprintf("%d: %v", pr.code, pr.err)
}

func NewMsgBrokerClient(host, port string) *KMsgBroker {
	kmb := &KMsgBroker{
		Host: host,
		Port: port,
		Writer: kafka.Writer{
			Addr:                   kafka.TCP(host + ":" + port), //127.0.0.1:9092 or kafka:9092 in docker
			AllowAutoTopicCreation: true,
		},
	}

	return kmb
}

func (kb *KMsgBroker) Publish(msg, key, topic string, resChan chan<- PublishRes) {
	messages := []kafka.Message{
		{
			Key:   []byte(key),
			Value: []byte(msg),
			Topic: topic,
		},
	}

	err := kb.Writer.WriteMessages(context.Background(), messages...)
	if err != nil {
		resChan <- PublishRes{code: 1, err: fmt.Errorf("failed to send message: %v", err)}
		return
	}

	log.Println(msg, "sent to", topic)
	resChan <- PublishRes{code: 0, err: nil}
}
