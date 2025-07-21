-- name: AddUserAuthority :exec
insert into user_authority(user_id, authority_id)
values ($1, $2);

-- name: DeleteUserAuthorities :exec
delete
from user_authority
where user_id = $1;

-- name: GetUserAuthorities :many
select a.id, a.authority
from authority a
         left join user_authority ua on ua.authority_id = a.id
where ua.user_id = $1
order by a.authority;