package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
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
			Content 		 string      `json:"content"`
			ReasoningContent string      `json:"reasoning_content"`
			Reasoning 		 string      `json:"reasoning"`
			ToolCalls        []ToolCall  `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type ToolCall struct {
	Id       string `json:"id"`
	Index    int    `json:"index"`
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

type MessageWrapper struct {
	Msg  Message
	Time string
}

// UI Events
type Event interface {
	Pretty() string
}

type ToolStart struct{ ID, Name, Args string }
func (t ToolStart) Pretty() string {
	return fmt.Sprintf("%s %s", t.Name, t.Args)
}


type ToolDone struct{ ID, Name, Result string }
func (t ToolDone) Pretty() string {
	return fmt.Sprintf("%s", t.Name)
}

type TurnDone struct{}


type Chunk struct {
    Data      string
    EOS       bool
    Reasoning bool
    TurnStart bool
}
