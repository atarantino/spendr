create table if not exists wallets (
    id serial primary key,
    name text not null,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);
