create table client_config (
	name text primary key,
	model text not null,
	message_context int not null
);

insert into client_config (name, model, message_context)
	WITH list(name) AS ( VALUES ('chat.message-context'))
	SELECT 'gpt-3.5-turbo', 'gpt-3.5-turbo', IFNULL(value, 5) 
	FROM list
	LEFT JOIN config USING (name);

insert into client_config (name, model, message_context) values ('gpt-4', 'gpt-4', 5);

delete from config where name = 'chat.message-context';
insert into config (name, value) values ('client-config', 'gpt-3.5-turbo');
