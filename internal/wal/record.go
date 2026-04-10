package wal

type OpType string

const (
	OpPut OpType = "put"
	OpDel OpType = "del"
)

type Record struct {
	Op    OpType `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"` // base64 encoded
}
