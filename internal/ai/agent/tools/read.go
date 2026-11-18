package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/thero-sgit/glyn/internal/common"
)

type ReadFile struct {
	Root string
}

func (r ReadFile) Run(ctx context.Context, args json.RawMessage) (string, error) {
	var rArgs readArgs
	if err := json.Unmarshal(args, &rArgs); err != nil {
		return "", err
	}

	//
	abs, err := common.SafePath(r.Root, rArgs.Path)
	if err != nil {
		return "", err
	}

	//
	file, err := os.Open(abs)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 
	f, err := common.BinaryCheck(file)
	if err != nil {
		return "", err
	}
	file = f

	//
	var builder strings.Builder
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	line, shown := 0, 0
	for scanner.Scan() {
		line++
		if line < rArgs.Offset {
			continue
		}
		
		if shown >= rArgs.Limit {
			fmt.Fprintf(
				&builder,
				"... truncated, file continues past line %d. Use offset=%d to read more.\n", line - 1, line,
			)

			break
		}

		text := scanner.Text()
		if len(text) > 500 {
			text = text[:500] + "..."
		}
		fmt.Fprintf(&builder, "%6d\t%s\n", line, text)
		shown++
	}

	return common.Truncate(builder.String(), 20_000), scanner.Err()
}

func (r ReadFile) Schema() map[string]any {
	return map[string]any{
		"type":      "function",
		"function" : map[string]any {
			"name":        "read_file",
			"description": "Read file of given path",

			"parameters":  map[string]any {
				"type": "object",

				"properties": map[string]any {
					"path": map[string]any {
						"type": "string",
						"description": "The path of file to read",
					},

					"offset": map[string]any {
						"type": "integer",
						"description": "line in file to start from (1-based start line)",
					},

					"limit": map[string]any {
						"type": "integer",
						"description": "line in file to stop at",
					},

				},

				"required": []string{"path", "offset", "limit"},
			},
		},
	}
}

func (r ReadFile) Name() string { return "read_file" }
// ---

type readArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}