package dal

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/gen/db/repository"
	"github.com/janobono/auth-service/pkg/util"
	"strings"
)

type SearchUsersParams struct {
	Page          int32
	Size          int32
	Sort          string
	SearchField   string
	Email         string
	AttributeKeys []string
}

func (q *Queries) CountUsersByCriteria(ctx context.Context, arg SearchUsersParams) (int64, error) {
	var query strings.Builder
	query.WriteString(`select count(*) from "user" u`)

	var conditions []string
	var parameters []interface{}
	paramIndex := 1

	if cond, params := buildEmailFilterCondition(arg.Email, &paramIndex); cond != "" {
		conditions = append(conditions, cond)
		parameters = append(parameters, params...)
	}

	searchValues := util.SplitWithoutBlank(util.ToScDf(arg.SearchField), " ")
	if len(searchValues) > 0 {
		if cond, params := buildEmailSearchConditions(searchValues, &paramIndex); cond != "" {
			conditions = append(conditions, cond)
			parameters = append(parameters, params...)
		}

		attributeKeys := util.Deduplicate(util.FilterBlank(arg.AttributeKeys))
		if len(attributeKeys) > 0 {
			joins, attrConditions, params := buildAttributeJoinsAndConditions(attributeKeys, searchValues, &paramIndex)
			query.WriteString(joins)
			conditions = append(conditions, attrConditions...)
			parameters = append(parameters, params...)
		}
	}

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	row := q.db.QueryRow(ctx, query.String(), parameters...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) SearchUsersByCriteria(ctx context.Context, arg SearchUsersParams) ([]repository.User, error) {
	var query strings.Builder
	query.WriteString(`select u.id, u.created_at, u.email, u.password, u.confirmed, u.enabled from "user" u`)

	var conditions []string
	var parameters []interface{}
	paramIndex := 1

	if cond, params := buildEmailFilterCondition(arg.Email, &paramIndex); cond != "" {
		conditions = append(conditions, cond)
		parameters = append(parameters, params...)
	}

	searchValues := util.SplitWithoutBlank(util.ToScDf(arg.SearchField), " ")
	if len(searchValues) > 0 {
		if cond, params := buildEmailSearchConditions(searchValues, &paramIndex); cond != "" {
			conditions = append(conditions, cond)
			parameters = append(parameters, params...)
		}

		attributeKeys := util.Deduplicate(util.FilterBlank(arg.AttributeKeys))
		if len(attributeKeys) > 0 {
			joins, attrConditions, params := buildAttributeJoinsAndConditions(attributeKeys, searchValues, &paramIndex)
			query.WriteString(joins)
			conditions = append(conditions, attrConditions...)
			parameters = append(parameters, params...)
		}
	}

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	query.WriteString(" order by " + arg.Sort)

	limit := arg.Size
	offset := arg.Page * limit
	query.WriteString(fmt.Sprintf(" limit %d offset %d", limit, offset))

	rows, err := q.db.Query(ctx, query.String(), parameters...)
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

func buildEmailFilterCondition(email string, paramIndex *int) (string, []interface{}) {
	if util.IsBlank(email) {
		return "", nil
	}
	cond := fmt.Sprintf("u.email like $%d", *paramIndex)
	param := "%" + util.ToScDf(email) + "%"
	*paramIndex++
	return cond, []interface{}{param}
}

func buildEmailSearchConditions(values []string, paramIndex *int) (string, []interface{}) {
	if len(values) == 0 {
		return "", nil
	}

	var sb strings.Builder
	params := make([]interface{}, 0, len(values))

	sb.WriteString("(")
	for i, val := range values {
		if i > 0 {
			sb.WriteString(" or ")
		}
		sb.WriteString(fmt.Sprintf("u.email like $%d", *paramIndex))
		params = append(params, "%"+val+"%")
		*paramIndex++
	}
	sb.WriteString(")")

	return sb.String(), params
}

func buildAttributeJoinsAndConditions(attributeKeys, values []string, paramIndex *int) (joins string, conditions []string, params []interface{}) {
	var joinBuilder strings.Builder
	conditions = []string{}
	params = []interface{}{}

	for i, key := range attributeKeys {
		ua := fmt.Sprintf("ua%d", i)
		a := fmt.Sprintf("a%d", i)

		joinBuilder.WriteString(fmt.Sprintf(`
join user_attribute %s on %s.user_id = u.id
join attribute %s on %s.id = %s.attribute_id
`, ua, ua, a, a, ua))

		var sb strings.Builder
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf("%s.key = $%d and (", a, *paramIndex))
		params = append(params, key)
		*paramIndex++

		for j, val := range values {
			if j > 0 {
				sb.WriteString(" or ")
			}
			sb.WriteString(fmt.Sprintf("unaccent(%s.value) ilike $%d", ua, *paramIndex))
			params = append(params, "%"+val+"%")
			*paramIndex++
		}

		sb.WriteString("))")
		conditions = append(conditions, sb.String())
	}

	return joinBuilder.String(), conditions, params
}
