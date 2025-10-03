create table if not exists transactions (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    plaid_account_id integer not null references plaid_accounts(id) on delete cascade,
    transaction_id text not null unique,
    account_id text not null,
    amount numeric(12,2) not null,
    date date not null,
    authorized_date date,
    name text not null,
    merchant_name text,
    pending boolean not null default false,
    payment_channel text not null,
    transaction_code text,
    iso_currency_code text,
    unofficial_currency_code text,
    location jsonb,
    payment_meta jsonb,
    personal_finance_category jsonb,
    counterparties jsonb,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

create index idx_transactions_user_id on transactions (user_id);
create index idx_transactions_plaid_account_id on transactions (plaid_account_id);
create unique index idx_transactions_transaction_id on transactions (transaction_id);
create index idx_transactions_date on transactions (date);
create index idx_transactions_pending on transactions (pending);
