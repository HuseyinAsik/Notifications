package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	writer *kafka.Writer
}

type Message struct {
	Key   []byte
	Value []byte
	Topic string
}

func NewWriter(brokers []string) *Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &Writer{
		writer: w,
	}
}

func (w *Writer) WriteMessages(
	ctx context.Context,
	messages []Message,
) error {

	var kafkaMessages []kafka.Message

	for _, m := range messages {
		kafkaMessages = append(kafkaMessages, kafka.Message{
			Key:   m.Key,
			Value: m.Value,
			Topic: m.Topic,
		})
	}

	return w.writer.WriteMessages(ctx, kafkaMessages...)
}
