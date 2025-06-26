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
	addData := repository.AddAttributeData{
		Key:      "test-key",
		Name:     "Test Name",
		Required: true,
		Hidden:   false,
	}

	createdAttr, err := repo.AddAttribute(ctx, addData)
	assert.NoError(t, err)
	assert.NotNil(t, createdAttr)
	assert.Equal(t, addData.Key, createdAttr.Key)

	// Get the attribute
	fetchedAttr, err := repo.GetAttribute(ctx, "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, fetchedAttr)
	assert.Equal(t, "Test Name", fetchedAttr.Name)

	// Delete the attribute
	err = repo.DeleteAttribute(ctx, createdAttr.ID)
	assert.NoError(t, err)

	// Try to get again (should fail)
	fetchedAttr, err = repo.GetAttribute(ctx, "test-key")
	assert.Error(t, err)
	assert.Nil(t, fetchedAttr)
}
