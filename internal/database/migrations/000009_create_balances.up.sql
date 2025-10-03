create table if not exists balances (
    wallet_id integer not null references wallets(id) on delete cascade,
    user_id integer not null references users(id) on delete cascade,
    net_balance numeric(12,2) not null default 0,
    last_updated_at timestamp default now() not null,
    primary key (wallet_id, user_id)
);

create index idx_balances_user_id on balances (user_id);
