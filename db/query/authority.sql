-- name: AddAuthority :one
insert into authority (id, authority)
values ($1, $2)
returning *;

-- name: CountAuthoritiesById :one
select count(*)
from authority
where id = $1;

-- name: CountAuthoritiesByAuthority :one
select count(*)
from authority
where authority = $1;

-- name: CountAuthoritiesByAuthorityNotId :one
select count(*)
from authority
where authority = $1
  and id != $2;

-- name: DeleteAuthorityById :exec
delete
from authority
where id = $1;

-- name: GetAllAuthorities :many
select *
from authority
order by authority;

-- name: GetAuthorityById :one
select *
from authority
where id = $1;

-- name: GetAuthorityByAuthority :one
select *
from authority
where authority = $1;

-- name: SetAuthority :one
update authority
set authority = $2
where id = $1
returning *;