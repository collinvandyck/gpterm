ALTER TABLE message RENAME TO message_old;
CREATE TABLE message (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	role text not null,
	content text not null
);
INSERT INTO message SELECT id, timestamp, role, content from message_old;
DROP TABLE message_old;

drop table conversation;

