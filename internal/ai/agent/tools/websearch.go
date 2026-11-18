package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/thero-sgit/glyn/internal/config"
)

type WebSearch struct {
	url string
	key string
}

func WebSearchTool() WebSearch {
	return WebSearch{
		url: "https://api.langsearch.com/v1/web-search",
		key: config.Config.LangSearchApiKey,
	}
}

func (w WebSearch) Run(ctx context.Context, args json.RawMessage) (string, error) {
	var searchArgs SearchRequest
	if err := json.Unmarshal(args, &searchArgs); err != nil {
		return "", err
	}

	response, err := w.getResults(ctx, searchArgs)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(response.Data, "", "	")
	if err != nil {
		return "", err
	}

	results := fmt.Sprintf("<search_results>\n%s\n<search_results>", data)
	return results, nil
}

func (w WebSearch) Schema() map[string]any {
	return map[string]any{
		"type":      "function",
		"function" : map[string]any {
			"name":        "web_search",
			"description": `
			Search the web with a query.

			Dates must be valid calender dates. A single date selects that UTC day; a range includes both endpoints and must have 'start <= end'
			Invalid dates or reversed dates return HTTP Status Code 400. Empty domain arrays apply no filtering.
			Use domain strings such as 'groq.com' and 'openai.com'.
			`,

			"parameters":  map[string]any {
				"type": "object",

				"properties": map[string]any {
					"query": map[string]any {
						"type": "string",
						"description": "Non-empty search query",
					},

					"count": map[string]any {
						"type": "integer",
						"description": "Request 1-50 results; values below 1 are normalized to 1; values above 50 return HTTP Status Code `400`. Default is 10",
					},

					"freshness": map[string]any {
						"type": "string",
						"description": "Results freshness. Default is `noLimit`",
						"enum": []string{
							"noLimit", "oneDay", "oneWeek", "oneMonth", "oneYear", "YYYY-MM-DD", "YYYY-MM-DD..YYYY-MM-DD",
						},
					},

					"includeDomains": map[string]any {
						"type": "array",
						"description": "List of domains to include. Restricts sources to these domains; entries must not be blank",
						"items": map[string]string {
							"type": "string",
						},
					},

					"excludeDomains": map[string]any {
						"type": "array",
						"description": "List of domains to exclude. Exclude these domains; entries must not be blank",
						"items": map[string]string {
							"type": "string",
						},
					},
				},

				"required": []string{"query"},
			},
		},
	}
}

func (w WebSearch) Name() string {return "web_search"}

func (w WebSearch) getResults(ctx context.Context, reqBody SearchRequest) (*SearchResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return &SearchResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return &SearchResponse{}, err
	}

	req.Header.Set("Authorization", "Bearer "+ w.key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &SearchResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &SearchResponse{}, fmt.Errorf("API Error (%d): %s", resp.StatusCode, string(body))
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return &SearchResponse{}, fmt.Errorf("Unmarshal Error %s:", err.Error())
	}

	return &response, nil
}

// ---
type SearchRequest struct {
	Query string            `json:"query"`
	Count uint8             `json:"count,omitempty"`
	Freshness string        `json:"freshness,omitempty"`
	IncludeDomains []string `json:"includeDomains,omitempty"`
	ExcludeDomains []string `json:"excludeDomains,omitempty"`
}

type SearchResult struct {
	Url string  	     `json:"url"`
	ID string   	     `json:"id,omitempty"`
	Name string 	     `json:"name,omitempty"`
	DisplayUrl string    `json:"displayUrl,omitempty"`
	Snippet string       `json:"snippet,omitempty"`
	Text string 	     `json:"text,omitempty"`
	DatePublished string `json:"datePublished,omitempty"`
}

type SearchResponse struct {
	Code int   `json:"code,omitempty"`
	LogID string `json:"log_id,omitempty"`
	Msg string 	 `json:"msg,omitempty"`
	Data struct {
		Type string `json:"_type"`

		QueryContext struct {
			OriginalQuery string `json:"originalQuery,omitempty"`
		} `json:"queryContext,omitempty"`

		WebPages struct {
			Value []SearchResult    `json:"value"`
			WebSearchUrl string     `json:"webSearchUrl,omitempty"`
			SomeResultsRemoved bool `json:"someResultsRemoved,omitempty"`
		} `json:"webPages,omitempty"`
	} `json:"data"`
	Usage struct {
		InputTokens  uint32 `json:"input_tokens"`
		OutputTokens uint32 `json:"output_tokens"`
	}`json:"usage"`
}