-- drop table if exists messages;
create table messages (
    id bigserial,
    txhash varchar(64) not null, 
    date timestamp with time zone not null,
    message_type varchar(50),
    sender varchar(100),
    recepient varchar(100),
    body jsonb
);