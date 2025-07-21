package repository_test

import (
	"context"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAttributeRepository_CRUD(t *testing.T) {
	repo := repository.NewAttributeRepository(DataSource)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add an attribute
	attribute, err := repo.AddAttribute(ctx, &repository.AttributeData{
		Key:      "test-key",
		Required: true,
		Hidden:   false,
	})
	assert.NoError(t, err)
	assert.NotNil(t, attribute)
	assert.Equal(t, "test-key", attribute.Key)
	assert.Equal(t, true, attribute.Required)
	assert.Equal(t, false, attribute.Hidden)

	// Count by id
	count, err := repo.CountById(ctx, attribute.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Count by key
	count, err = repo.CountByKey(ctx, attribute.Key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Count by key and not id
	count, err = repo.CountByKeyAndNotId(ctx, attribute.Key, attribute.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Get the attribute by key
	fetched, err := repo.GetAttributeByKey(ctx, attribute.Key)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, attribute.ID, fetched.ID)
	assert.Equal(t, attribute.Required, fetched.Required)
	assert.Equal(t, attribute.Hidden, fetched.Hidden)

	// Set the attribute
	changed, err := repo.SetAttribute(ctx, attribute.ID, &repository.AttributeData{
		Key:      attribute.Key,
		Required: !attribute.Required,
		Hidden:   !attribute.Hidden,
	})
	assert.NoError(t, err)
	assert.NotNil(t, changed)
	assert.Equal(t, attribute.ID, changed.ID)
	assert.Equal(t, attribute.Key, changed.Key)
	assert.Equal(t, !attribute.Required, changed.Required)
	assert.Equal(t, !attribute.Hidden, changed.Hidden)

	// Get the attribute by id
	fetched, err = repo.GetAttributeById(ctx, changed.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, changed.Key, fetched.Key)
	assert.Equal(t, changed.Required, fetched.Required)
	assert.Equal(t, changed.Hidden, fetched.Hidden)

	// Search attribute
	_, err = repo.AddAttribute(ctx, &repository.AttributeData{
		Key:      "other-key",
		Required: true,
		Hidden:   false,
	})
	assert.NoError(t, err)

	page, err := repo.SearchAttributes(ctx,
		&repository.SearchAttributesCriteria{SearchField: attribute.Key},
		&common.Pageable{
			Page: 0,
			Size: 10,
			Sort: "id",
		})
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, false, page.Empty)
	assert.Equal(t, true, page.First)
	assert.Equal(t, true, page.Last)
	assert.Equal(t, int32(1), page.TotalPages)
	assert.Equal(t, int64(1), page.TotalElements)
	assert.Equal(t, 1, len(page.Content))

	page, err = repo.SearchAttributes(ctx,
		&repository.SearchAttributesCriteria{},
		&common.Pageable{
			Page: 1,
			Size: 1,
			Sort: "id",
		})
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, false, page.Empty)
	assert.Equal(t, false, page.First)
	assert.Equal(t, true, page.Last)
	assert.Equal(t, int32(2), page.TotalPages)
	assert.Equal(t, int64(2), page.TotalElements)
	assert.Equal(t, 1, len(page.Content))

	// Delete the attribute
	err = repo.DeleteAttributeById(ctx, attribute.ID)
	assert.NoError(t, err)

	// Try to get again (should fail)
	fetched, err = repo.GetAttributeById(ctx, attribute.ID)
	assert.Error(t, err)
	assert.Nil(t, fetched)
}
