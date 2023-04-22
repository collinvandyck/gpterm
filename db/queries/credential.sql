-- name: GetCredential :one
SELECT value FROM credential
WHERE name=? LIMIT 1;

-- name: UpdateCredential :exec
INSERT OR REPLACE INTO credential (name, value) VALUES (?, ?);

