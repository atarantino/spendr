create table if not exists transaction_categorizations (
    id serial primary key,
    transaction_id integer not null references transactions(id) on delete cascade,
    wallet_id integer not null references wallets(id) on delete cascade,
    category_type text not null check (category_type in ('shared', 'individual')),
    categorized_by_user_id integer not null references users(id) on delete cascade,
    categorized_at timestamp default now() not null,
    unique (transaction_id, wallet_id)
);

create index idx_transaction_categorizations_transaction_id on transaction_categorizations (transaction_id);
create index idx_transaction_categorizations_wallet_id on transaction_categorizations (wallet_id);
create index idx_transaction_categorizations_category_type on transaction_categorizations (category_type);
