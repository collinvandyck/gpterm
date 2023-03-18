create table message (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	role text not null,
	content text not null
)
