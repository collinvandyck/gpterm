CREATE TABLE credential (
	name varchar primary key,
	value text NOT NULL
);
CREATE TABLE message (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	role text not null,
	content text not null
);
CREATE TABLE usage (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	prompt_tokens integer not null,
	completion_tokens integer not null,
	total_tokens integer not null
);
