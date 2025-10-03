create table if not exists plaid_accounts (
    id serial primary key,
    plaid_item_id integer not null references plaid_items(id) on delete cascade,
    account_id text not null unique,
    name text not null,
    official_name text,
    type text not null,
    subtype text,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

create index idx_plaid_accounts_plaid_item_id on plaid_accounts (plaid_item_id);
create unique index idx_plaid_accounts_account_id on plaid_accounts (account_id);
