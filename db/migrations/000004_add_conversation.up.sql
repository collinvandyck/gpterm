create table conversation (
	id integer primary key,
	name text,
	protected integer not null default 0,
	selected integer not null default 0
);
insert into conversation(id, name, protected, selected) values (0, 'default', true, true);

ALTER TABLE message RENAME TO message_old;
CREATE TABLE message (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	role text not null,
	content text not null, 
	conversation_id integer not null default 0,
	FOREIGN KEY (conversation_id) REFERENCES conversation(id)
);
INSERT INTO message SELECT id, timestamp, role, content, 0 from message_old;
DROP TABLE message_old;
