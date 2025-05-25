-- name: AddJwk :one
insert into jwk (id, kty, use, alg, public_key, private_key, active, created_at, expires_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *;

-- name: DeleteNotActiveJwks :exec
delete
from jwk
where active is false
  and use = $1
  and expires_at < $2;

-- name: GetJwk :one
select *
from jwk
where id = $1
limit 1;

-- name: GetActiveJwk :one
select *
from jwk
where active is true
  and use = $1
order by created_at
limit 1;

-- name: DeactivateJwks :exec
update jwk
set active = false
where id != $1
  and use = $2;
