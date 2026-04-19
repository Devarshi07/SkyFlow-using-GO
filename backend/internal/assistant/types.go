package assistant

import (
	"time"
)

// Conversation represents a chat session stored in Redis
type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	Role       string     `json:"role"` // "user" | "assistant" | "system" | "tool"
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Timestamp  time.Time  `json:"timestamp"`
}

// ToolCall represents a tool call made by the assistant
type ToolCall struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Args     string    `json:"arguments"` // JSON string of arguments
	Result   string    `json:"result,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Groq chat completion types
type groqRequest struct {
	Model       string     `json:"model"`
	Messages    []groqMsg  `json:"messages"`
	Tools       []groqTool `json:"tools,omitempty"`
	ToolChoice  string     `json:"tool_choice,omitempty"`
	Temperature float64    `json:"temperature"`
	MaxTokens   int        `json:"max_tokens,omitempty"`
}

type groqToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type groqToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function groqToolCallFunc `json:"function"`
}

type groqMsg struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []groqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type groqTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  any    `json:"parameters"`
	} `json:"function"`
}

type groqResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message groqMsg `json:"message"`
	} `json:"choices"`
}