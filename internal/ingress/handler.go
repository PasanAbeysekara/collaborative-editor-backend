package ingress

import (
	"context"
	"time"

	"yourmodule/internal/collab"
	"yourmodule/internal/kafkabus"
)

type Handler struct {
	bus *kafkabus.Bus
}

func NewHandler(bus *kafkabus.Bus) *Handler { return &Handler{bus: bus} }

// Example: call from your WS "on message" or HTTP POST controller
func (h *Handler) OnClientEdit(op collab.ClientOp) {
	// 1) (Optional) echo back to connected clients immediately for snappy UI
	//    Your existing WS broadcast line can stay here if you want.

	// 2) Durable enqueue to Kafka
	_ = h.bus.PublishClientOp(context.Background(), op)
}

// Example bootstrapping (called from your main.go)
func InitIngress() *Handler {
	return NewHandler(kafkabus.New())
}

func (h *Handler) Close() { _ = h.bus.Close() }

// A tiny example of usage (remove in real app):
func (h *Handler) demo() {
	h.OnClientEdit(collab.ClientOp{
		OpID:       "uuid-1",
		DocumentID: "doc_123",
		UserID:     "u_1",
		Type:       "insert",
		Payload:    map[string]interface{}{"pos": 0, "text": "H"},
		ClientTs:   time.Now().UnixMilli(),
	})
}
