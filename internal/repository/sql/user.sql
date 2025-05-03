-- name: GetUser :one
select *
from sa_user
where id = $1 limit 1;

-- name: InsertUser :one
insert into sa_user (email, password, first_name, last_name, confirmed, enabled)
values ($1, $2, $3, $4, $5, $6) RETURNING *;