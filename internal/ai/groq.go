package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/config"
)

type StreamRequest struct {
	Model           string           `json:"model"`
	Messages        []agent.Message  `json:"messages"`
	ReasoningFormat string           `json:"reasoning_format,omitempty"`
	ReasoningEffort string           `json:"reasoning_effort,omitempty"`
	Stream          bool             `json:"stream"`
	Tools           []map[string]any `json:"tools,omitempty"`
}

type groq struct {
	url    string
	apiKey string
}

func newGroq() *groq {
	return &groq{
		url:    "https://api.groq.com/openai/v1",
		apiKey: config.Config.GroqApiKey,
	}
}

func (g groq) CreateChatCompletionStream(ctx context.Context, reqBody StreamRequest) (*agent.ChatCompletionStream, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return &agent.ChatCompletionStream{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.url+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return &agent.ChatCompletionStream{}, err
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &agent.ChatCompletionStream{}, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &agent.ChatCompletionStream{}, fmt.Errorf("API Error (%d): %s", resp.StatusCode, string(body))
	}

	return &agent.ChatCompletionStream{
		Body:   resp.Body,
		Reader: bufio.NewReader(resp.Body),
	}, nil
}

func (g groq) handlePrompt(ctx context.Context, a *agent.Agent, messages []agent.Message) (agent.ChatCompletionStream, error) {
	req := StreamRequest{
		Model:           config.Config.Model.Name,
		Messages:        messages,
		ReasoningEffort: config.Config.Effort.Name,
		ReasoningFormat: "parsed",
		Stream:          true,
		Tools:           a.ToolSchemas(),
	}

	stream, err := g.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return agent.ChatCompletionStream{}, fmt.Errorf("groq completion failed: %w", err)
	}

	return *stream, nil
}

type GroqModelInfo struct {
	Id                          string   `json:"id"`
	Object                      string   `json:"object"`
	Created                     int      `json:"created"`
	OwnedBy                     string   `json:"owned_by"`
	Active                      bool     `json:"active"`
	ContextWindow               int      `json:"context_window"`
	PublicApps                  string   `json:"public_apps,omitempty"`
	MaxCompletionTokens         int      `json:"max_completion_tokens"`
	HuggingFaceId               string   `json:"hugging_face_id"`
	Name                        string   `json:"name"`
	InputModalities             []string `json:"input_modalities"`
	OutputModalities            []string `json:"output_modalities"`
	ContextLength               int      `json:"context_length"`
	MaxOutputLength             int      `json:"max_output_length"`
	SupportedSamplingParameters []string `json:"supported_sampling_parameters"`
	SupportedFeatures           []string `json:"supported_features"`
}

type GroqModelList struct {
	Object string          `json:"object"`
	Data   []GroqModelInfo `json:"data"`
}

func (g groq) getModels(ctx context.Context) (GroqModelList, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.url+"/models", nil)
	if err != nil {
		return GroqModelList{}, err
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	const retryInterval = 3 * time.Second
	for {
		resp, err := http.DefaultClient.Do(req.Clone(ctx))
		if err != nil {
			if ctx.Err() != nil {
				return GroqModelList{}, ctx.Err()
			}
			if err := waitForModelRetry(ctx, retryInterval); err != nil {
				return GroqModelList{}, err
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if ctx.Err() != nil {
				return GroqModelList{}, ctx.Err()
			}
			if err := waitForModelRetry(ctx, retryInterval); err != nil {
				return GroqModelList{}, err
			}
			continue
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			err := fmt.Errorf("models API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < http.StatusInternalServerError {
				return GroqModelList{}, err
			}
			if waitErr := waitForModelRetry(ctx, retryInterval); waitErr != nil {
				return GroqModelList{}, waitErr
			}
			continue
		}

		var modelList GroqModelList
		if err := json.Unmarshal(body, &modelList); err != nil {
			return GroqModelList{}, fmt.Errorf("decode models response: %w", err)
		}
		return modelList, nil
	}
}

func waitForModelRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
