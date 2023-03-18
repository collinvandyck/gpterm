-- name: GetMessages :many
SELECT * FROM message;

-- name: InsertMessage :exec
INSERT INTO message (timestamp, role, content) 
VALUES (?,?,?);
