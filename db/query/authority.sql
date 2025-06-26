-- name: AddAuthority :one
insert into authority (id, authority)
values ($1, $2)
returning *;

-- name: DeleteAuthority :exec
delete
from authority
where id = $1;

-- name: GetAuthority :one
select *
from authority
where authority = $1
limit 1;

-- name: GetUserAuthorities :many
select a.id, a.authority
from authority a
         left join user_authority ua on ua.authority_id = a.id
where ua.user_id = $1
order by a.authority;