create extension if not exists unaccent;

create table sa_authority
(
    id        bigint generated always as identity primary key,
    authority varchar(255) not null unique
);

create table sa_user
(
    id         bigint generated always as identity primary key,
    email      varchar(255) not null unique,
    password   varchar(255) not null,
    first_name varchar(255) not null,
    last_name  varchar(255) not null,
    confirmed  bool         not null,
    enabled    bool         not null
);

create table sa_user_authority
(
    user_id      bigint not null references sa_user (id) on delete cascade,
    authority_id bigint not null references sa_authority (id) on delete cascade,
    primary key (user_id, authority_id)
);