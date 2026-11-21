package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Tool interface {
	Name() string
	Schema() map[string]any
	Run(ctx context.Context, args json.RawMessage) (string, error)
}

func WrapToolResult(name, result string) string {
	return fmt.Sprintf("<tool_results:%s>\n%s\n<tool_results>", name, result)
}

func TrimWrap(prefix, s string) (string, bool) {
	opening := prefix
	if !strings.HasPrefix(s, opening) {
		return s, false
	}

	const closing = "<tool_results>"
	if !strings.HasSuffix(s, closing) {
		return s, false
	}

	result := strings.TrimPrefix(s, opening)
	result = strings.TrimSuffix(result, closing)
	result = strings.TrimPrefix(result, "\n")
	result = strings.TrimSuffix(result, "\n")
	return result, true
}
