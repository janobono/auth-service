package repository_test

import (
	"context"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAuthorityRepository_CRUD(t *testing.T) {
	repo := repository.NewAuthorityRepository(DataSource)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add an authority
	addData := repository.AddAuthorityData{
		Authority: "ROLE_TEST",
	}

	created, err := repo.AddAuthority(ctx, addData)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "ROLE_TEST", created.Authority)

	// Get the authority
	fetched, err := repo.GetAuthority(ctx, "ROLE_TEST")
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)

	// Delete the authority
	err = repo.DeleteAuthority(ctx, created.ID)
	assert.NoError(t, err)

	// Ensure it’s deleted
	_, err = repo.GetAuthority(ctx, "ROLE_TEST")
	assert.Error(t, err)
}
