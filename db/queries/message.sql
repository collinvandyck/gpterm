-- name: GetMessages :many
SELECT * FROM message;

-- name: GetLatestMessages :many
select *
from message
where id in (select id from message order by id desc limit ?)
order by id;

-- name: InsertMessage :exec
INSERT INTO message (role, content) 
VALUES (?,?);
