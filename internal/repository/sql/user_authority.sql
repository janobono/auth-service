-- name: AddUserAuthority :exec
insert into sa_user_authority(user_id, authority_id)
values ($1, $2);

-- name: DeleteUserAuthorities :exec
delete
from sa_user_authority
where user_id = $1;

-- name: GetUserAuthorities :many
select *
from sa_user_authority
where user_id = $1;