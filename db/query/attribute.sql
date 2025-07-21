-- name: AddAttribute :one
insert into attribute (id, key, required, hidden)
values ($1, $2, $3, $4)
returning *;

-- name: CountAttributesById :one
select count(*)
from attribute
where id = $1;

-- name: CountAttributesByKey :one
select count(*)
from attribute
where key = $1;

-- name: CountAttributesByKeyNotId :one
select count(*)
from attribute
where key = $1
  and id != $2;

-- name: DeleteAttributeById :exec
delete
from attribute
where id = $1;

-- name: GetAttributeById :one
select *
from attribute
where id = $1;

-- name: GetAttributeByKey :one
select *
from attribute
where key = $1;

-- name: SetAttribute :one
update attribute
set key      = $2,
    required = $3,
    hidden   = $4
where id = $1
returning *;