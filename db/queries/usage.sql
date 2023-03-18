-- name: InsertUsage :exec
INSERT INTO usage 
(prompt_tokens, completion_tokens, total_tokens)
VALUES
(?,?,?);

-- name: GetTotalTokens :one
SELECT total(total_tokens) as integer from usage;

-- name: GetCompletionTokens :one
SELECT total(completion_tokens) from usage;

-- name: GetPromptTokens :one
SELECT total(prompt_tokens) from usage;


