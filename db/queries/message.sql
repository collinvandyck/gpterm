-- name: GetMessages :many
SELECT * FROM message;

-- name: InsertMessage :exec
INSERT INTO message (role, content) 
VALUES (?,?);
