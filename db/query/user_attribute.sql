-- name: AddUserAttribute :exec
insert into user_attribute(user_id, attribute_id, value)
values ($1, $2, $3);

-- name: DeleteUserAttributes :exec
delete
from user_attribute
where user_id = $1;

-- name: GetUserAttributes :many
select a.id, a.key, ua.value, a.required, a.hidden
from attribute a
         left join user_attribute ua on ua.attribute_id = a.id
where ua.user_id = $1
order by a.key;