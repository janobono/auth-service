-- name: AddUser :one
insert into sa_user (email, password, first_name, last_name, confirmed, enabled)
values ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: CountUsers :one
select count(*)
from sa_user;

-- name: CountUsersByCriteria :one
select count(*)
from sa_user
where ($1 is null or (
    (email like '%' || $1 || '%') or
    (unaccent(first_name) ILIKE '%' || $1 || '%') or
    (unaccent(last_name) ILIKE '%' || $1 || '%')
    ))
  and ($2 is null or email = $2)
;

-- name: DeleteUser :exec
delete
from sa_user
where id = $1;

-- name: GetUser :one
select *
from sa_user
where id = $1 limit 1;

-- name: SearchUsersByCriteria :many
select *
from sa_user
where ($1 is null or (
    (email like '%' || $1 || '%') or
    (unaccent(first_name) ILIKE '%' || $1 || '%') or
    (unaccent(last_name) ILIKE '%' || $1 || '%')
    ))
  and ($2 is null or email = $2)
order by case
             when $3 = 'email' then email
             when $3 = 'first_name' then first_name
             when $3 = 'last_name' then last_name
             else id
    end
       , case
             when $4 = 'desc' then 1
             else 0
    end
    desc limit $5
offset $6;