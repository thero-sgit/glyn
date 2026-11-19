package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/thero-sgit/glyn/internal/ai/common"
	co "github.com/thero-sgit/glyn/internal/common"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ai/agent/tools"
)

type Agent struct {
	role     string
	maxTurns int
	Tools    map[string]tools.Tool
}

func NewAgent(role string, maxTurns int) Agent {
	tools := map[string]tools.Tool{
		"read_file": tools.ReadFile{
			Root: config.Config.WorkingDir,
		},
		"web_search": tools.WebSearchTool(),
	}

	return Agent{
		role:     role,
		maxTurns: maxTurns,
		Tools:    tools,
	}
}

func (a *Agent) Run(
	ctx context.Context,
	streamFunc func(ctx context.Context, a *Agent, messages []common.Message) (common.ChatCompletionStream, error),
	history []common.Message,
	messageTs string,
	c chan common.Chunk,
	eventsChan chan<- common.Event,
	wtdc chan<- common.MessageWrapper,
) ([]common.Message, error) {
	for turn := 0; turn < a.maxTurns; turn++ {
		if err := ctx.Err(); err != nil {
			return history, err
		}

		msg, finish, err := a.streamTurn(ctx, streamFunc, history, c, wtdc, messageTs)
		now := time.Now().Format(co.TsFormat)
		if err != nil {
			return history, err
		}

		if msg.Role == "tool" {
			history = common.PruneAndFit(history)
		}

		history = append(history, msg)
		persistMessage(ctx, wtdc, common.MessageWrapper{Msg: msg, Time: now})

		if len(msg.ToolCalls) == 0 && finish == "stop" {
			return history, nil
		}
		if len(msg.ToolCalls) == 0 {
			continue
		}

		results := a.executeTools(ctx, msg.ToolCalls, eventsChan)
		history = append(history, results...)
		for _, result := range results {
			persistMessage(ctx, wtdc, common.MessageWrapper{
				Msg:  result,
				Time: time.Now().Format(co.TsFormat),
			})
		}

		if ctx.Err() != nil {
			return history, ctx.Err()
		}
	}

	return history, errors.New("max turns reached")
}

func (a *Agent) streamTurn(
	ctx context.Context,
	streamFunc func(ctx context.Context, a *Agent, messages []common.Message) (common.ChatCompletionStream, error),
	history []common.Message,
	c chan common.Chunk,
	wtdc chan<- common.MessageWrapper,
	messageTs string,
) (common.Message, string, error) {

	stream, err := streamFunc(ctx, a, history)
	if err != nil {
		return common.Message{}, "", err
	}
	defer stream.Close()
	if err := sendChunk(ctx, c, common.Chunk{TurnStart: true}); err != nil {
		return common.Message{}, "", err
	}

	if msg := history[len(history)-1]; msg.Role == "user" {
		persistMessage(ctx, wtdc, common.MessageWrapper{Msg: msg, Time: messageTs})
	}

	var text strings.Builder
	var calls []common.ToolCall
	var finish string

	for {
		select {
		case <-ctx.Done():
			return common.Message{Role: "assistant", Content: text.String(), ToolCalls: calls}, finish, nil
		default:
		}

		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return common.Message{}, "", err
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
			if err := sendChunk(ctx, c, common.Chunk{Data: reasoningContent, Reasoning: true}); err != nil {
				return common.Message{}, "", err
			}
		}

		content := delta.Content
		if content != "" {
			if err := sendChunk(ctx, c, common.Chunk{Data: content}); err != nil {
				return common.Message{}, "", err
			}
			text.WriteString(content)
		}

		for _, tc := range delta.ToolCalls {
			for len(calls) <= tc.Index {
				calls = append(calls, common.ToolCall{Type: "function"})
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
			if finish == "stop" {
				if err := sendChunk(ctx, c, common.Chunk{EOS: true}); err != nil {
					return common.Message{}, "", err
				}
			}
			break
		}
	}

	return common.Message{Role: "assistant", Content: text.String(), ToolCalls: calls}, finish, nil
}

func persistMessage(ctx context.Context, wtdc chan<- common.MessageWrapper, msg common.MessageWrapper) {
	select {
	case wtdc <- msg:
	case <-ctx.Done():
	}
}

func sendChunk(ctx context.Context, chunks chan<- common.Chunk, chunk common.Chunk) error {
	select {
	case chunks <- chunk:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *Agent) executeTools(ctx context.Context, calls []common.ToolCall, eventsChan chan<- common.Event) []common.Message {
	outputs := make([]common.Message, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case eventsChan <- common.ToolStart{ID: call.Id, Name: call.Function.Name, Args: call.Function.Arguments}:
			case <-ctx.Done():
				return
			}

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
					result = r
				}
			}

			outputs[i] = common.Message{Role: "tool", ToolCallId: call.Id, Content: result}
			select {
			case eventsChan <- common.ToolDone{
				ID: call.Id,
				Name: call.Function.Name,
				Result: result,
			}:
			case <-ctx.Done():
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
