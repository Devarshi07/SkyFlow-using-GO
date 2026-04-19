package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

const groqBaseURL = "https://api.groq.com/openai/v1"

type GroqClient struct {
	apiKey   string
	model    string
	baseURL  string
	httpClient *http.Client
}

func NewGroqClient() *GroqClient {
	model := os.Getenv("GROQ_MODEL")
	if model == "" {
		model = "llama-3.1-8b-instant"
	}
	return &GroqClient{
		apiKey:     os.Getenv("GROQ_API_KEY"),
		model:      model,
		baseURL:    groqBaseURL,
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
}

var retryAfterRe = regexp.MustCompile(`try again in ([0-9]+(?:\.[0-9]+)?)s`)

// Chat sends a chat completion request to Groq. On 429 it retries once after
// the server-suggested delay (capped at 6s) before giving up.
func (c *GroqClient) Chat(ctx context.Context, msgs []groqMsg, tools []groqTool) (*groqResponse, error) {
	reqPayload := groqRequest{
		Model:       c.model,
		Messages:    msgs,
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 0.2,
		MaxTokens:   1024,
	}
	if len(tools) == 0 {
		reqPayload.ToolChoice = ""
	}
	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to call Groq API: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			var groqResp groqResponse
			decErr := json.NewDecoder(resp.Body).Decode(&groqResp)
			resp.Body.Close()
			if decErr != nil {
				return nil, fmt.Errorf("failed to decode Groq response: %w", decErr)
			}
			return &groqResp, nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			wait := parseRetryAfter(resp.Header.Get("Retry-After"), string(respBody))
			if wait > 6*time.Second {
				wait = 6 * time.Second
			}
			select {
			case <-time.After(wait):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return nil, fmt.Errorf("Groq API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil, fmt.Errorf("Groq API: exhausted retries")
}

func parseRetryAfter(header, body string) time.Duration {
	if header != "" {
		if secs, err := strconv.ParseFloat(header, 64); err == nil && secs > 0 {
			return time.Duration(secs * float64(time.Second))
		}
	}
	if m := retryAfterRe.FindStringSubmatch(body); len(m) == 2 {
		if secs, err := strconv.ParseFloat(m[1], 64); err == nil && secs > 0 {
			return time.Duration(secs * float64(time.Second))
		}
	}
	return 2 * time.Second
}