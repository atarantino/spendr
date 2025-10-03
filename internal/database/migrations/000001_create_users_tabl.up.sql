create table if not exists users (
    id serial primary key,
    name varchar(255) not null,
    email varchar(255) not null,
    password_hash varchar(255) not null,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null
);

create unique index idx_users_email on users (email);
