package main

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

// TODO: wire these to your actual implementations.
func LoadDocState(documentID string) (any, error) { return nil, nil }
func ApplyOp(state any, op collab.ClientOp) (any, error) { return state, nil }
func SaveDocState(documentID string, newState any, opID string) error { return nil }

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	topic := os.Getenv("TOPIC_DOC_EDITS")
	if topic == "" {
		topic = "doc-edits.v1"
	}
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "collab-workers"
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(brokers, ","),
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,          // keep defaults simple
		MaxBytes:       10e6,
		CommitInterval: time.Second, // autocommit every second
	})
	defer r.Close()

	log.Printf("collab-worker started: brokers=%s topic=%s group=%s", brokers, topic, groupID)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("kafka read error: %v", err)
		}

		var op collab.ClientOp
		if err := json.Unmarshal(m.Value, &op); err != nil {
			log.Printf("skip invalid message: %v", err)
			continue
		}

		// --- Your existing pipeline ---
		state, err := LoadDocState(op.DocumentID)
		if err != nil {
			log.Printf("load state failed: %v", err)
			continue
		}

		newState, err := ApplyOp(state, op)
		if err != nil {
			log.Printf("apply op failed: %v", err)
			continue
		}

		if err := SaveDocState(op.DocumentID, newState, op.OpID); err != nil {
			log.Printf("save state failed: %v", err)
			continue
		}

		// Optional: broadcast validated op to clients here
		// BroadcastValidated(op, newState)
	}
}
