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

type AttributeRepository interface {
	AddAttribute(ctx context.Context, data *AttributeData) (*Attribute, error)
	CountById(ctx context.Context, id pgtype.UUID) (int64, error)
	CountByKey(ctx context.Context, key string) (int64, error)
	CountByKeyAndNotId(ctx context.Context, key string, id pgtype.UUID) (int64, error)
	DeleteAttributeById(ctx context.Context, id pgtype.UUID) error
	GetAllAttributes(ctx context.Context) ([]*Attribute, error)
	GetAttributeById(ctx context.Context, id pgtype.UUID) (*Attribute, error)
	GetAttributeByKey(ctx context.Context, key string) (*Attribute, error)
	SearchAttributes(ctx context.Context, criteria *SearchAttributesCriteria, pageable *common.Pageable) (*common.Page[*Attribute], error)
	SetAttribute(ctx context.Context, id pgtype.UUID, data *AttributeData) (*Attribute, error)
}

type attributeRepositoryImpl struct {
	dataSource *db.DataSource
}

func NewAttributeRepository(dataSource *db.DataSource) AttributeRepository {
	return &attributeRepositoryImpl{dataSource}
}

func (a *attributeRepositoryImpl) AddAttribute(ctx context.Context, data *AttributeData) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.AddAttribute(ctx, sqlc.AddAttributeParams{
		ID:       db2.NewUUID(),
		Key:      data.Key,
		Required: data.Required,
		Hidden:   data.Hidden,
	})

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

func (a *attributeRepositoryImpl) CountById(ctx context.Context, id pgtype.UUID) (int64, error) {
	return a.dataSource.Queries.CountAttributesById(ctx, id)
}

func (a *attributeRepositoryImpl) CountByKey(ctx context.Context, key string) (int64, error) {
	return a.dataSource.Queries.CountAttributesByKey(ctx, key)
}

func (a *attributeRepositoryImpl) CountByKeyAndNotId(ctx context.Context, key string, id pgtype.UUID) (int64, error) {
	return a.dataSource.Queries.CountAttributesByKeyNotId(ctx, sqlc.CountAttributesByKeyNotIdParams{
		Key: key,
		ID:  id,
	})
}

func (a *attributeRepositoryImpl) DeleteAttributeById(ctx context.Context, id pgtype.UUID) error {
	return a.dataSource.Queries.DeleteAttributeById(ctx, id)
}

func (a *attributeRepositoryImpl) GetAllAttributes(ctx context.Context) ([]*Attribute, error) {
	attributes, err := a.dataSource.Queries.GetAllAttributes(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*Attribute, len(attributes))
	for index, attribute := range attributes {
		result[index] = toAttribute(&attribute)
	}
	return result, nil
}

func (a *attributeRepositoryImpl) GetAttributeById(ctx context.Context, id pgtype.UUID) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.GetAttributeById(ctx, id)

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

func (a *attributeRepositoryImpl) GetAttributeByKey(ctx context.Context, key string) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.GetAttributeByKey(ctx, key)

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

func (a *attributeRepositoryImpl) SearchAttributes(ctx context.Context, criteria *SearchAttributesCriteria, pageable *common.Pageable) (*common.Page[*Attribute], error) {
	totalRows, err := a.countAttributes(ctx, criteria)

	if err != nil {
		return nil, err
	}

	content, err := a.searchAttributes(ctx, criteria, pageable)

	if err != nil {
		return nil, err
	}

	return common.NewPage[*Attribute](pageable, totalRows, content), nil
}

func (a *attributeRepositoryImpl) SetAttribute(ctx context.Context, id pgtype.UUID, data *AttributeData) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.SetAttribute(ctx, sqlc.SetAttributeParams{
		ID:       id,
		Key:      data.Key,
		Required: data.Required,
		Hidden:   data.Hidden,
	})

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

func (a *attributeRepositoryImpl) countAttributes(ctx context.Context, criteria *SearchAttributesCriteria) (int64, error) {
	var query strings.Builder
	query.WriteString("select count(*) from attribute a")

	paramIndex := 1
	conditions, parameters := a.buildSearchQueryParts(criteria, &paramIndex)

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	row := a.dataSource.Pool.QueryRow(ctx, query.String(), parameters...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (a *attributeRepositoryImpl) searchAttributes(ctx context.Context, criteria *SearchAttributesCriteria, pageable *common.Pageable) ([]*Attribute, error) {
	var query strings.Builder
	query.WriteString("select a.id, a.key, a.required, a.hidden from attribute a")

	paramIndex := 1
	conditions, parameters := a.buildSearchQueryParts(criteria, &paramIndex)

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	query.WriteString(" order by " + pageable.Sort)
	query.WriteString(fmt.Sprintf(" limit %d offset %d", pageable.Limit(), pageable.Offset()))

	rows, err := a.dataSource.Pool.Query(ctx, query.String(), parameters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []*Attribute
	for rows.Next() {
		var attribute sqlc.Attribute
		if err := rows.Scan(
			&attribute.ID,
			&attribute.Key,
			&attribute.Required,
			&attribute.Hidden,
		); err != nil {
			return nil, err
		}
		content = append(content, toAttribute(&attribute))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return content, nil
}

func (a *attributeRepositoryImpl) buildSearchQueryParts(criteria *SearchAttributesCriteria, paramIndex *int) (conditions []string, parameters []interface{}) {
	conditions = []string{}
	parameters = []interface{}{}

	searchValues := common.SplitWithoutBlank(common.ToScDf(criteria.SearchField), " ")
	if len(searchValues) > 0 {
		if cond, params := a.buildKeySearchConditions(searchValues, paramIndex); cond != "" {
			conditions = append(conditions, cond)
			parameters = append(parameters, params...)
		}
	}

	return conditions, parameters
}

func (a *attributeRepositoryImpl) buildKeySearchConditions(values []string, paramIndex *int) (string, []interface{}) {
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
		sb.WriteString(fmt.Sprintf("unaccent(a.key) ilike $%d", *paramIndex))
		params = append(params, "%"+val+"%")
		*paramIndex++
	}
	sb.WriteString(")")

	return sb.String(), params
}
