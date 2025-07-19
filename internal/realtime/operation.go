package realtime

type OpType string

const (
	OpInsert OpType = "insert"
	OpDelete OpType = "delete"
)

type Operation struct {
	Type    OpType `json:"type"`
	Pos     int    `json:"pos"`     // Position of the change
	Text    string `json:"text"`    // Text to insert (for insert ops)
	Len     int    `json:"len"`     // Number of characters to delete (for delete ops)
	Version int    `json:"version"` // Document version this op is based on
}
