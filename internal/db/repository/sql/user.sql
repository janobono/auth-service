-- name: AddUser :one
insert into sa_user (email, password, first_name, last_name, confirmed, enabled)
values ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: CountUsers :one
select count(*)
from sa_user;

-- name: CountUsersByCriteria :one
select count(*)
from sa_user
where ($1 = '' or (
    (email like '%' || $1 || '%') or
    (unaccent(first_name) ILIKE '%' || $1 || '%') or
    (unaccent(last_name) ILIKE '%' || $1 || '%')
    ))
  and ($2 = '' or email = $2)
;

-- name: DeleteUser :exec
delete
from sa_user
where id = $1;

-- name: GetUser :one
select *
from sa_user
where id = $1 limit 1;

-- name: GetUserByEmail :one
select *
from sa_user
where email = $1 limit 1;

-- name: SearchUsersByCriteria :many
select *
from sa_user
where ($1 = '' or (
    (email like '%' || $1 || '%') or
    (unaccent(first_name) ILIKE '%' || $1 || '%') or
    (unaccent(last_name) ILIKE '%' || $1 || '%')
    ))
  and ($2 = '' or email = $2)
order by $3 limit $4
offset $5;

-- name: SetUser :exec
update sa_user
set email      = $2,
    password   = $3,
    first_name = $4,
    last_name  = $5,
    confirmed  = $6,
    enabled    = $7
where id = $1;