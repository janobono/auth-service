-- name: AddAttribute :one
insert into attribute (id, key, required, hidden)
values ($1, $2, $3, $4)
returning *;

-- name: DeleteAttribute :exec
delete
from attribute
where id = $1;

-- name: GetAttribute :one
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