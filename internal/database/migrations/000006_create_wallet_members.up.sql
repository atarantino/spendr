create table if not exists wallet_members (
    wallet_id integer not null references wallets(id) on delete cascade,
    user_id integer not null references users(id) on delete cascade,
    joined_at timestamp default now() not null,
    primary key (wallet_id, user_id)
);

create index idx_wallet_members_user_id on wallet_members (user_id);
