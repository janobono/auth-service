package dal

import (
	"context"
	"github.com/janobono/auth-service/gen/db/repository"
)

type SearchUsersParams struct {
	Page          int32
	Size          int32
	Sort          string
	SearchField   string
	Email         string
	AttributeKeys []string
}

const countUsersByCriteria = `-- name: CountAllUsers :one
select count(*) from "user"
`

func (q *Queries) CountUsersByCriteria(ctx context.Context, arg SearchUsersParams) (int64, error) {
	var parameters []interface{}

	// TODO implement

	row := q.db.QueryRow(ctx, countUsersByCriteria, parameters...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const searchUsersByCriteria = `-- name: SearchUsersByCriteria :many
select id, created_at, email, password, confirmed, enabled
from "user"
where id = $1
limit 1
`

func (q *Queries) SearchUsersByCriteria(ctx context.Context, arg SearchUsersParams) ([]repository.User, error) {
	var parameters []interface{}

	// TODO implement

	rows, err := q.db.Query(ctx, searchUsersByCriteria, parameters...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repository.User
	for rows.Next() {
		var i repository.User
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.Email,
			&i.Password,
			&i.Confirmed,
			&i.Enabled,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
