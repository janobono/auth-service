package repository_test

import (
	"context"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwkRepository_CRUD(t *testing.T) {
	repo := repository.NewJwkRepository(DataSource)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add new JWK
	addData := repository.JwkData{
		Use:        "sig",
		Expiration: 2 * time.Hour,
	}
	created, err := repo.AddJwk(ctx, addData)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "sig", created.Use)
	assert.True(t, created.Active)

	// Get active JWK
	active, err := repo.GetActiveJwk(ctx, "sig")
	assert.NoError(t, err)
	assert.NotNil(t, active)
	assert.Equal(t, created.ID, active.ID)

	// Get active JWKs
	activeJwks, err := repo.GetActiveJwks(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, activeJwks)
	assert.Equal(t, created.ID, activeJwks[0].ID)

	// Get JWK by ID
	fetched, err := repo.GetJwk(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "RSA", fetched.Kty)
}
