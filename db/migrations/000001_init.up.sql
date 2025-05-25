create extension if not exists unaccent;

create table if not exists jwk
(
    id          uuid primary key,
    kty         varchar(255) not null,
    use         varchar(255) not null,
    alg         varchar(255) not null,
    public_key  bytea        not null,
    private_key bytea        not null,
    active      boolean      not null,
    created_at  timestamptz  not null,
    expires_at  timestamptz  not null
);

create table if not exists attribute
(
    id       uuid primary key,
    key      varchar(255) not null,
    name     varchar(255) not null,
    required boolean      not null,
    hidden   boolean      not null
);

create table if not exists authority
(
    id        uuid primary key,
    authority varchar(255) not null unique
);

create table if not exists "user"
(
    id         uuid primary key,
    created_at timestamptz  not null,
    email      varchar(255) not null unique,
    password   varchar(255) not null,
    confirmed  bool         not null,
    enabled    bool         not null
);

create table if not exists user_attribute
(
    user_id      uuid         not null references "user" (id) on delete cascade,
    attribute_id uuid         not null references attribute (id) on delete cascade,
    value        varchar(255) not null,
    primary key (user_id, attribute_id)
);

create table if not exists user_authority
(
    user_id      uuid not null references "user" (id) on delete cascade,
    authority_id uuid not null references authority (id) on delete cascade,
    primary key (user_id, authority_id)
);
