-- name: GetConversations :many
SELECT * FROM conversation order by id;

-- name: ConversationCount :one
select count(*) from conversation;

-- name: GetActiveConversation :one
select * from conversation where selected=true;

-- name: CreateConversation :one
insert into conversation (name) values (null)
returning *;

-- name: UnsetSelectedConversation :exec
update conversation
set selected = false;

-- name: SetSelectedConversation :exec
update conversation
set selected = true
where id=?;

-- name: NextConversation :one
select * from conversation
where id > (
	select id from conversation where selected = true
)
order by id
limit 1;

-- name: PreviousConversation :one
select * from conversation
where id < (
	select id from conversation where selected = true
)
order by id desc
limit 1;

-- name: DeleteConversation :one
delete from conversation where id = ? returning *;
