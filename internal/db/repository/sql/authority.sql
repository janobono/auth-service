-- name: AddAuthority :one
insert into sa_authority (authority)
values ($1) RETURNING *;

-- name: GetAuthority :one
select *
from sa_authority
where authority = $1 limit 1;

-- name: GetAuthorities :many
select *
from sa_authority
order by authority;

-- name: GetUserAuthorities :many
select a.id, a.authority
from sa_authority a
         left join sa_user_authority ua on ua.authority_id = a.id
where ua.user_id = $1
order by a.authority;