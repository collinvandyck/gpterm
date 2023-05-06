CREATE TABLE credential (
	name varchar primary key,
	value text NOT NULL
);
CREATE TABLE usage (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	prompt_tokens integer not null,
	completion_tokens integer not null,
	total_tokens integer not null
);
CREATE TABLE conversation (
	id integer primary key,
	name text,
	protected integer not null default 0,
	selected integer not null default 0
);
CREATE TABLE message (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	role text not null,
	content text not null, 
	conversation_id integer not null default 0,
	FOREIGN KEY (conversation_id) REFERENCES conversation(id)
);
CREATE INDEX message_conversation_id on message (conversation_id);
CREATE TABLE config (
	name varchar primary key,
	value text NOT NULL
);
CREATE TABLE client_config (
	name text primary key,
	model text not null,
	message_context int not null
);
