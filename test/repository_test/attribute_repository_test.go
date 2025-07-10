package repository_test

import (
	"context"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAttributeRepository_CRUD(t *testing.T) {
	repo := repository.NewAttributeRepository(DataSource)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create an attribute
	addData := repository.AttributeData{
		Key:      "test-key",
		Required: true,
		Hidden:   false,
	}

	createdAttr, err := repo.AddAttribute(ctx, addData)
	assert.NoError(t, err)
	assert.NotNil(t, createdAttr)
	assert.Equal(t, addData.Key, createdAttr.Key)

	// Count by id
	count, err := repo.CountById(ctx, createdAttr.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Count by key
	count, err = repo.CountByKey(ctx, createdAttr.Key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Get the attribute
	fetchedAttr, err := repo.GetAttribute(ctx, "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, fetchedAttr)
	assert.Equal(t, true, fetchedAttr.Required)
	assert.Equal(t, false, fetchedAttr.Hidden)

	// Set the attribute
	setData := repository.AttributeData{
		Key:      fetchedAttr.Key,
		Required: !fetchedAttr.Required,
		Hidden:   !fetchedAttr.Hidden,
	}

	changedAttr, err := repo.SetAttribute(ctx, fetchedAttr.ID, setData)
	assert.NoError(t, err)
	assert.NotNil(t, changedAttr)
	assert.Equal(t, fetchedAttr.Key, changedAttr.Key)
	assert.Equal(t, !fetchedAttr.Required, changedAttr.Required)
	assert.Equal(t, !fetchedAttr.Hidden, changedAttr.Hidden)

	// Delete the attribute
	err = repo.DeleteAttribute(ctx, createdAttr.ID)
	assert.NoError(t, err)

	// Try to get again (should fail)
	fetchedAttr, err = repo.GetAttribute(ctx, "test-key")
	assert.Error(t, err)
	assert.Nil(t, fetchedAttr)
}
