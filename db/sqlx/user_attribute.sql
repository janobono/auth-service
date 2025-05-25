-- name: AddUserAttribute :exec
insert into user_attribute(user_id, attribute_id, value)
values ($1, $2, $3);

-- name: DeleteUserAttributes :exec
delete
from user_attribute
where user_id = $1;
