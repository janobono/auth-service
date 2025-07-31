-- name: AddUser :one
insert into "user" (id, created_at, email, password, confirmed, enabled)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: CountAllUsers :one
select count(*)
from "user";

-- name: CountUsersById :one
select count(*)
from "user"
where id = $1;

-- name: CountUsersByEmail :one
select count(*)
from "user"
where email = $1;

-- name: CountUsersByEmailNotId :one
select count(*)
from "user"
where email = $1
  and id != $2;

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

-- name: SetUserConfirmed :one
update "user"
set confirmed = $2
where id = $1
returning *;

-- name: SetUserEmail :one
update "user"
set email = $2
where id = $1
returning *;

-- name: SetUserEnabled :one
update "user"
set enabled = $2
where id = $1
returning *;