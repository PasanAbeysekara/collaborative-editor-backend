package kafkabus

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"yourmodule/internal/collab"
)

type Bus struct {
	writer *kafka.Writer
}

func New() *Bus {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	topic := os.Getenv("TOPIC_DOC_EDITS")
	if topic == "" {
		topic = "doc-edits.v1"
	}

	return &Bus{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(strings.Split(brokers, ",")...),
			Topic:        topic,
			Balancer:     &kafka.Hash{}, // key-based (documentId) partitioning
			RequiredAcks: kafka.RequireAll,
			Async:        false,         // keep it simple for assignment
		},
	}
}

func (b *Bus) Close() error {
	return b.writer.Close()
}

// PublishClientOp is called from your WS/API ingress to enqueue the edit.
func (b *Bus) PublishClientOp(ctx context.Context, op collab.ClientOp) error {
	// small timeout so ingress never blocks forever
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	payload, err := json.Marshal(op)
	if err != nil {
		return err
	}
	return b.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(op.DocumentID), // keep all ops of a doc ordered
		Value: payload,
	})
}

// Helper (optional): simple logger for success testing
func (b *Bus) MustPublish(ctx context.Context, op collab.ClientOp) {
	if err := b.PublishClientOp(ctx, op); err != nil {
		log.Printf("kafka publish failed: %v", err)
	}
}
