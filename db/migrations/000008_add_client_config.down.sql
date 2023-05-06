drop table client_config;
insert into config (name,value) values ('chat.message-context', 5);
delete from config where name = 'client-config';
