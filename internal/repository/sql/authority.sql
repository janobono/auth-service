-- name: GetAuthority :one
select * from sa_authority
where authority = $1 limit 1;

-- name: ListAuthorities :many
select * from sa_authority
order by authority;