-- name: AddAttribute :one
insert into attribute (id, key, required, hidden)
values ($1, $2, $3, $4)
returning *;

-- name: CountAttributeById :one
select count(*)
from attribute
where id = $1;

-- name: CountAttributeByKey :one
select count(*)
from attribute
where key = $1;

-- name: CountAttributeByKeyNotId :one
select count(*)
from attribute
where key = $1
  and id != $2;

-- name: DeleteAttribute :exec
delete
from attribute
where id = $1;

-- name: GetAttributeById :one
select *
from attribute
where id = $1
limit 1;

-- name: GetAttributeByKey :one
select *
from attribute
where key = $1
limit 1;

-- name: GetUserAttributes :many
select a.id, a.key, ua.value, a.required, a.hidden
from attribute a
         left join user_attribute ua on ua.attribute_id = a.id
where ua.user_id = $1
order by a.key;

-- name: SetAttribute :one
update attribute
set key      = $2,
    required = $3,
    hidden   = $4
where id = $1
returning *;