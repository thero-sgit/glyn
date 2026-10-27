package agent

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
