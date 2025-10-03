create table if not exists plaid_items (
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    access_token text not null,
    item_id text not null unique,
    institution_name text,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

create index idx_plaid_items_user_id on plaid_items (user_id);
create unique index idx_plaid_items_item_id on plaid_items (item_id);
