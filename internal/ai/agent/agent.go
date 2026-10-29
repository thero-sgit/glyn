package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"github.com/thero-sgit/glyn/internal/common"
	"github.com/thero-sgit/glyn/internal/config"
)

type Agent struct {
	role     string
	maxTurns int
	Tools    map[string]Tool
}

func NewAgent(role string, maxTurns int) Agent {
    tools := map[string]Tool {
		"read_file":  ReadFile{
			config.Config.WorkingDir,
		},
	}

	return Agent{
		role:     role,
		maxTurns: maxTurns,
		Tools: tools,
	}
}

func (a *Agent) Run(
	ctx context.Context, 
	streamFunc func(ctx context.Context, a *Agent, messages []Message) (ChatCompletionStream, error), 
	history []Message, 
	c chan Chunk, 
	eventsChan chan<- Events,
) ([]Message, error) {
	for turn := 0; turn < a.maxTurns; turn++ {
		msg, finish, err := a.streamTurn(ctx, streamFunc, history, c, eventsChan)
		if err != nil {
			return history, err
		}

		history = append(history, msg)

		if len(msg.ToolCalls) == 0 && finish != "tool_calls" {
			return history, nil
		}	

		//
		results := a.executeTools(ctx, msg.ToolCalls, eventsChan)
		history = append(history, results...)

		if ctx.Err() != nil {
            return history, ctx.Err()
        }
	}

	return history, errors.New("max turns reached")
}

func (a *Agent) streamTurn(
	ctx context.Context,
	streamFunc func(ctx context.Context, a *Agent, messages []Message) (ChatCompletionStream, error),
	history []Message,
	c chan Chunk,
	eventsChan chan<- Events,
) (Message, string, error) {

	stream, err := streamFunc(ctx, a, history)
	if err != nil {
		return Message{}, "", err
	}
	defer stream.Close()

	var text   strings.Builder
	var calls  []ToolCall
	var finish string

	for {
		select {
		case <-ctx.Done():
			return Message{ Role: "assistant", Content: text.String(), ToolCalls: calls}, finish, nil
		default:
		}

		chunk, err := stream.Recv()
		if err != nil {
			break
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		reasoningContent := delta.ReasoningContent
		if reasoningContent == "" {
			reasoningContent = delta.Reasoning
		}

		if reasoningContent != "" {
			c <- Chunk{Data: reasoningContent, EOS: false, Reasoning: true}
			continue
		}

		content := delta.Content
		c <- Chunk{Data: content, EOS: false}
		text.WriteString(content)

		for _, tc := range delta.ToolCalls {
			for len(calls) <= tc.Index {
				calls = append(calls, ToolCall{Type: "function"})
			}

			call := &calls[tc.Index]
			if tc.Id != "" {
				call.Id = tc.Id
			}
			if tc.Function.Name != "" {
				call.Function.Name = tc.Function.Name
			}

			call.Function.Arguments = call.Function.Arguments + tc.Function.Arguments
		}

		if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != "" {
			finish = *chunk.Choices[0].FinishReason
			break
		}
	}

	return Message{ Role: "assistant", Content: text.String(), ToolCalls: calls}, finish, nil
}

func (a *Agent) executeTools(ctx context.Context, calls []ToolCall, eventsChan chan<- Events) []Message {
	outputs := make([]Message, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// eventsChan <- ToolStart{call.Id, call.Function.Name, call.Function.Arguments}

			var result string
			tool, ok := a.Tools[call.Function.Name]

			switch {
			case !ok:
				result = fmt.Sprintf("error: unknown tool %s", call.Function.Name)
			case !json.Valid([]byte(call.Function.Arguments)):
				result = "error: arguments were not valid JSON, please retry"
			default:
				r, err := tool.Run(ctx, json.RawMessage(call.Function.Arguments))
				if err != nil {
					result = "error: " + err.Error()
				} else {
					result = common.Truncate(r, 20_000)
				}

				// eventsChan <- ToolDone{
				// 	call.Id,
				// 	call.Function.Name,
				// 	result,
				// }

				outputs[i] = Message{Role: "tool", ToolCallId: call.Id, Content: result}
			}
		}()
	}

	wg.Wait()
	return outputs
}

func (a *Agent) ToolSchemas() []map[string]any {
	var schemas []map[string]any
	for _, tool := range a.Tools {
		schemas = append(schemas, tool.Schema())
	}

	return schemas
}