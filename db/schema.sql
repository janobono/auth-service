-- Enable extension
create extension if not exists unaccent;

-- Table: jwk
create table if not exists jwk
(
    id          uuid         not null,
    kty         varchar(255) not null,
    use         varchar(255) not null,
    alg         varchar(255) not null,
    public_key  bytea        not null,
    private_key bytea        not null,
    active      boolean      not null,
    created_at  timestamptz  not null,
    expires_at  timestamptz  not null
);

alter table jwk
    add constraint pk_jwk primary key (id);

-- Table: attribute
create table if not exists attribute
(
    id       uuid         not null,
    key      varchar(255) not null,
    required boolean      not null,
    hidden   boolean      not null
);

alter table attribute
    add constraint pk_attribute primary key (id);

-- Table: authority
create table if not exists authority
(
    id        uuid         not null,
    authority varchar(255) not null
);

alter table authority
    add constraint pk_authority primary key (id);

alter table authority
    add constraint uq_authority_authority unique (authority);

-- Table: user
create table if not exists "user"
(
    id         uuid         not null,
    created_at timestamptz  not null,
    email      varchar(255) not null,
    password   varchar(255) not null,
    confirmed  bool         not null,
    enabled    bool         not null
);

alter table "user"
    add constraint pk_user primary key (id);

alter table "user"
    add constraint uq_user_email unique (email);

-- Table: user_attribute
create table if not exists user_attribute
(
    user_id      uuid         not null,
    attribute_id uuid         not null,
    value        varchar(255) not null
);

alter table user_attribute
    add constraint pk_user_attribute primary key (user_id, attribute_id);

alter table user_attribute
    add constraint fk_user_attribute_user foreign key (user_id) references "user" (id) on delete cascade;

alter table user_attribute
    add constraint fk_user_attribute_attribute foreign key (attribute_id) references attribute (id) on delete cascade;

-- Table: user_authority
create table if not exists user_authority
(
    user_id      uuid not null,
    authority_id uuid not null
);

alter table user_authority
    add constraint pk_user_authority primary key (user_id, authority_id);

alter table user_authority
    add constraint fk_user_authority_user foreign key (user_id) references "user" (id) on delete cascade;

alter table user_authority
    add constraint fk_user_authority_authority foreign key (authority_id) references authority (id) on delete cascade;
