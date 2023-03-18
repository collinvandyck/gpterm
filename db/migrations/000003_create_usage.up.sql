create table usage (
	id integer primary key,
	timestamp datetime not null default current_timestamp,
	prompt_tokens integer not null,
	completion_tokens integer not null,
	total_tokens integer not null
)
