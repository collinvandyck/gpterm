CREATE TABLE credential (
	name varchar primary key,
	value text NOT NULL
);
CREATE TABLE message (
	id integer primary key,
	timestamp text,
	role text,
	content text
);
