-- name: GetAPIKey :one
SELECT value FROM credential
WHERE name='api_key' LIMIT 1;

-- name: InsertAPIKey :exec
INSERT INTO credential (name, value) values ('api_key', ?);

-- name: UpdateAPIKey :exec
UPDATE credential set value=? where name='api_key';

