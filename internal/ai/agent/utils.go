package agent

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type ChatCompletionStream struct {
	Body   io.ReadCloser
	Reader *bufio.Reader
}

func (s *ChatCompletionStream) Close() {
	s.Body.Close()
}

func (s *ChatCompletionStream) Recv() (GroqStreamChunk, error) {
	line, err := s.Reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return GroqStreamChunk{}, err
		}
	}

	line = strings.TrimSpace(line)

	if !strings.HasPrefix(line, "data: ") {
		return GroqStreamChunk{}, nil
	}

	data := strings.TrimPrefix(line, "data: ")

	if data == "[DONE]" {
		return GroqStreamChunk{}, nil
	}

	var chunk GroqStreamChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return GroqStreamChunk{}, err
	}

	return chunk, nil
}

type GroqStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content 		 string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			Reasoning 		 string `json:"reasoning"`
		} `json:"delta"`
	} `json:"choices"`
}

type ToolCall struct {
	Index    int    `json:"index"`
	Id       string `json:"id"`
	Type     string `json:"type"`

	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Message struct {
	Role 	   string      `json:"role"`
	Content    string      `json:"content,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallId string      `json:"tool_call_id,omitempty"`
}

// UI Events
type Events interface {}
type TextDelta  struct{ Text string }
type ToolStart  struct{ ID, Name, Args string }
type ToolDone   struct{ ID, Name, Result string }
type TurnDone   struct{}
