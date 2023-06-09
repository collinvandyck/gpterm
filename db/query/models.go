// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2

package query

import (
	"database/sql"
	"time"
)

type ClientConfig struct {
	Name           string `json:"name"`
	Model          string `json:"model"`
	MessageContext int64  `json:"message_context"`
}

type Config struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Conversation struct {
	ID        int64          `json:"id"`
	Name      sql.NullString `json:"name"`
	Protected int64          `json:"protected"`
	Selected  int64          `json:"selected"`
}

type Credential struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Message struct {
	ID             int64     `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	ConversationID int64     `json:"conversation_id"`
}

type Usage struct {
	ID               int64     `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	PromptTokens     int64     `json:"prompt_tokens"`
	CompletionTokens int64     `json:"completion_tokens"`
	TotalTokens      int64     `json:"total_tokens"`
}
