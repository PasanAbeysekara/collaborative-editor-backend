package realtime

type OpType string

type ServerMessageType string

const (
	MsgInitialState ServerMessageType = "initial_state"
	MsgOperation    ServerMessageType = "operation"
)

// wrapper for all messages sent to clients.
type ServerMessage struct {
	Type    ServerMessageType `json:"type"`
	Content string            `json:"content,omitempty"` // For initial state
	Version int               `json:"version,omitempty"` // <<< ADD THIS LINE
	Op      *Operation        `json:"op,omitempty"`      // For operations
}

const (
	OpInsert OpType = "insert"
	OpDelete OpType = "delete"
	OpUndo   OpType = "undo"
)

type Operation struct {
	Type    OpType `json:"type"`
	Pos     int    `json:"pos"`     // Position of the change
	Text    string `json:"text"`    // Text to insert (for insert ops)
	Len     int    `json:"len"`     // Number of characters to delete (for delete ops)
	Version int    `json:"version"` // Document version this op is based on
}
