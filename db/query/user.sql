-- name: AddUser :one
insert into "user" (id, created_at, email, password, confirmed, enabled)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: CountAllUsers :one
select count(*) from "user";

-- name: DeleteUserById :exec
delete
from "user"
where id = $1;

-- name: GetUserById :one
select *
from "user"
where id = $1
limit 1;

-- name: GetUserByEmail :one
select *
from "user"
where email = $1
limit 1;