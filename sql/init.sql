create table transactions (
    txhash varchar(64) not null UNIQUE,
    txdate timestamp with time zone not null,
    fcdtxid bigint not null UNIQUE,
    txdata jsonb
);

create index idx_txdate on transactions(txdate);
