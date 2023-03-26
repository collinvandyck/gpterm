-- name: CountMessagesForConversation :one
select count(*) 
from message
where conversation_id = ?;

-- name: GetMessages :many
SELECT * FROM message;

-- name: GetLatestMessages :many
select *
from message
where id in (
	select m.id 
	from message m 
	join conversation c on m.conversation_id = c.id
	where c.selected = true
	order by m.id desc 
	limit ?
)
order by id;

-- name: GetPreviousMessageForRole :one
select m.*
from message m
join conversation c on m.conversation_id = c.id
where m.role = ?
and c.selected = true
order by m.id desc
limit 1 offset ?
;

-- name: InsertMessage :exec
INSERT INTO message (role, content, conversation_id) 
SELECT ?, ?, id
from conversation
where selected = true
;
