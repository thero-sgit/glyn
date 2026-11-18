package tools

import (
	"context"
	"encoding/json"
)

type Tool interface {
	Name()   string
	Schema() map[string]any
	Run(ctx context.Context, args json.RawMessage) (string, error)
}