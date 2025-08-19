package collab

// ClientOp is the tiny JSON we pass through Kafka.
type ClientOp struct {
	OpID       string                 `json:"opId"`
	DocumentID string                 `json:"documentId"`
	UserID     string                 `json:"userId"`
	Type       string                 `json:"type"`    // "insert" | "delete" | "cursor" | "selection"
	Payload    map[string]interface{} `json:"payload"` // your op content
	ClientTs   int64                  `json:"clientTs"`
}