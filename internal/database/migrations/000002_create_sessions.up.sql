create table if not exists sessions (
    token text primary key,
    data bytea not null,
    expiry timestamptz not null
);

create index idx_sessions_expiry on sessions (expiry);
