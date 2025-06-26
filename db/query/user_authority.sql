-- name: AddUserAuthority :exec
insert into user_authority(user_id, authority_id)
values ($1, $2);

-- name: DeleteUserAuthorities :exec
delete
from user_authority
where user_id = $1;
