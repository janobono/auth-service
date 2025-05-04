-- name: AddAuthority :one
insert into sa_authority (authority)
values ($1) RETURNING *;

-- name: DeleteAuthority :exec
delete
from sa_authority
where id = $1;

-- name: GetAuthority :one
select *
from sa_authority
where authority = $1 limit 1;

-- name: ListAuthorities :many
select *
from sa_authority
order by authority;