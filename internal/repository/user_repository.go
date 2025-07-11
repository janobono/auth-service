package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/sqlc"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/go-util/common"
	db2 "github.com/janobono/go-util/db"
	"strings"
)

type UserRepository interface {
	AddUser(ctx context.Context, data UserData) (*User, error)
	DeleteUser(ctx context.Context, id pgtype.UUID) error
	GetUserAttributes(ctx context.Context, userID pgtype.UUID) ([]*UserAttribute, error)
	GetUserAuthorities(ctx context.Context, userID pgtype.UUID) ([]*Authority, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserById(ctx context.Context, id pgtype.UUID) (*User, error)
	SearchUsers(ctx context.Context, criteria *SearchUsersCriteria, pageable *common.Pageable) (*common.Page[*User], error)
	SetUserAttributes(ctx context.Context, arg UserAttributesData) ([]*UserAttribute, error)
	SetUserAuthorities(ctx context.Context, arg UserAuthoritiesData) ([]*Authority, error)
}

type userRepositoryImpl struct {
	dataSource *db.DataSource
}

func NewUserRepository(dataSource *db.DataSource) UserRepository {
	return &userRepositoryImpl{dataSource}
}

func (u *userRepositoryImpl) AddUser(ctx context.Context, data UserData) (*User, error) {
	user, err := u.dataSource.Queries.AddUser(ctx, sqlc.AddUserParams{
		ID:        db2.NewUUID(),
		CreatedAt: db2.NowUTC(),
		Email:     data.Email,
		Password:  data.Password,
		Enabled:   data.Enabled,
		Confirmed: data.Confirmed,
	})

	if err != nil {
		return nil, err
	}

	return toUser(&user), nil
}

func (u *userRepositoryImpl) DeleteUser(ctx context.Context, id pgtype.UUID) error {
	return u.dataSource.Queries.DeleteUser(ctx, id)
}

func (u *userRepositoryImpl) GetUserAttributes(ctx context.Context, userID pgtype.UUID) ([]*UserAttribute, error) {
	var result []*UserAttribute

	userAttributes, err := u.dataSource.Queries.GetUserAttributes(ctx, userID)

	if err != nil {
		return result, err
	}

	for _, userAttribute := range userAttributes {
		result = append(result, toUserAttribute(&userAttribute))
	}

	return result, nil
}

func (u *userRepositoryImpl) GetUserAuthorities(ctx context.Context, userID pgtype.UUID) ([]*Authority, error) {
	var result []*Authority

	userAuthorities, err := u.dataSource.Queries.GetUserAuthorities(ctx, userID)

	if err != nil {
		return result, err
	}

	for _, userAuthority := range userAuthorities {
		result = append(result, toAuthority(&userAuthority))
	}

	return result, nil
}

func (u *userRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := u.dataSource.Queries.GetUserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	return toUser(&user), nil
}

func (u *userRepositoryImpl) GetUserById(ctx context.Context, id pgtype.UUID) (*User, error) {
	user, err := u.dataSource.Queries.GetUserById(ctx, id)

	if err != nil {
		return nil, err
	}

	return toUser(&user), nil
}

func (u *userRepositoryImpl) SearchUsers(ctx context.Context, criteria *SearchUsersCriteria, pageable *common.Pageable) (*common.Page[*User], error) {
	totalRows, err := u.countUsers(ctx, criteria)

	if err != nil {
		return nil, err
	}

	content, err := u.searchUsers(ctx, criteria, pageable)

	if err != nil {
		return nil, err
	}

	return common.NewPage[*User](pageable, totalRows, content), nil
}

func (u *userRepositoryImpl) SetUserAttributes(ctx context.Context, arg UserAttributesData) ([]*UserAttribute, error) {
	_, err := u.dataSource.ExecTx(ctx, func(q *sqlc.Queries) (interface{}, error) {
		err := q.DeleteUserAttributes(ctx, arg.UserID)

		if err != nil {
			return nil, err
		}

		for _, attribute := range arg.Attributes {
			err = q.AddUserAttribute(ctx, sqlc.AddUserAttributeParams{
				UserID:      arg.UserID,
				AttributeID: attribute.Attribute.ID,
				Value:       attribute.Value,
			})

			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return arg.Attributes, nil
}

func (u *userRepositoryImpl) SetUserAuthorities(ctx context.Context, arg UserAuthoritiesData) ([]*Authority, error) {
	_, err := u.dataSource.ExecTx(ctx, func(q *sqlc.Queries) (interface{}, error) {
		err := q.DeleteUserAuthorities(ctx, arg.UserID)

		if err != nil {
			return nil, err
		}

		for _, authority := range arg.Authorities {
			err = q.AddUserAuthority(ctx, sqlc.AddUserAuthorityParams{
				UserID:      arg.UserID,
				AuthorityID: authority.ID,
			})

			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return arg.Authorities, nil
}

func (u *userRepositoryImpl) countUsers(ctx context.Context, criteria *SearchUsersCriteria) (int64, error) {
	var query strings.Builder
	query.WriteString(`select count(*) from "user" u`)

	paramIndex := 1
	joins, conditions, parameters := u.buildSearchQueryParts(criteria, &paramIndex)
	query.WriteString(joins)

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	row := u.dataSource.Pool.QueryRow(ctx, query.String(), parameters...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (u *userRepositoryImpl) searchUsers(ctx context.Context, criteria *SearchUsersCriteria, pageable *common.Pageable) ([]*User, error) {
	var query strings.Builder
	query.WriteString(`select u.id, u.created_at, u.email, u.password, u.confirmed, u.enabled from "user" u`)

	paramIndex := 1
	joins, conditions, parameters := u.buildSearchQueryParts(criteria, &paramIndex)
	query.WriteString(joins)

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	query.WriteString(" order by " + pageable.Sort)
	query.WriteString(fmt.Sprintf(" limit %d offset %d", pageable.Limit(), pageable.Offset()))

	rows, err := u.dataSource.Pool.Query(ctx, query.String(), parameters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []*User
	for rows.Next() {
		var user sqlc.User
		if err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.Email,
			&user.Password,
			&user.Confirmed,
			&user.Enabled,
		); err != nil {
			return nil, err
		}
		content = append(content, toUser(&user))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return content, nil
}

func (u *userRepositoryImpl) buildSearchQueryParts(criteria *SearchUsersCriteria, paramIndex *int) (joins string, conditions []string, parameters []interface{}) {
	var joinBuilder strings.Builder
	conditions = []string{}
	parameters = []interface{}{}

	if cond, params := u.buildEmailFilterCondition(criteria.Email, paramIndex); cond != "" {
		conditions = append(conditions, cond)
		parameters = append(parameters, params...)
	}

	searchValues := common.SplitWithoutBlank(common.ToScDf(criteria.SearchField), " ")
	if len(searchValues) > 0 {
		if cond, params := u.buildEmailSearchConditions(searchValues, paramIndex); cond != "" {
			conditions = append(conditions, cond)
			parameters = append(parameters, params...)
		}

		attributeKeys := common.Deduplicate(common.FilterBlank(criteria.AttributeKeys))
		if len(attributeKeys) > 0 {
			joins, attrConditions, params := u.buildAttributeJoinsAndConditions(attributeKeys, searchValues, paramIndex)
			joinBuilder.WriteString(joins)
			conditions = append(conditions, attrConditions...)
			parameters = append(parameters, params...)
		}
	}

	return joinBuilder.String(), conditions, parameters
}

func (u *userRepositoryImpl) buildEmailFilterCondition(email string, paramIndex *int) (string, []interface{}) {
	if common.IsBlank(email) {
		return "", nil
	}
	cond := fmt.Sprintf("u.email like $%d", *paramIndex)
	param := "%" + common.ToScDf(email) + "%"
	*paramIndex++
	return cond, []interface{}{param}
}

func (u *userRepositoryImpl) buildEmailSearchConditions(values []string, paramIndex *int) (string, []interface{}) {
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

func (u *userRepositoryImpl) buildAttributeJoinsAndConditions(attributeKeys, values []string, paramIndex *int) (joins string, conditions []string, params []interface{}) {
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
