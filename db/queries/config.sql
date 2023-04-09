-- name: GetConfig :many
SELECT * FROM config;

-- name: GetConfigValue :one
SELECT value FROM config
WHERE name=?;

-- name: SetConfigValue :exec
INSERT OR REPLACE INTO config (name, value) VALUES (?, ?);

