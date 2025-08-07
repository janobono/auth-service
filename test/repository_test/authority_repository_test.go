package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/stretchr/testify/assert"
)

func TestAuthorityRepository_CRUD(t *testing.T) {
	repo := repository.NewAuthorityRepository(DataSource)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add an authority
	authority, err := repo.AddAuthority(ctx, &repository.AuthorityData{
		Authority: "ROLE_TEST",
	})
	assert.NoError(t, err)
	assert.NotNil(t, authority)
	assert.Equal(t, "ROLE_TEST", authority.Authority)

	// Count by id
	count, err := repo.CountById(ctx, authority.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Count by authority
	count, err = repo.CountByAuthority(ctx, authority.Authority)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Count by authority and not id
	count, err = repo.CountByAuthorityAndNotId(ctx, authority.Authority, authority.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Get the authority by authority
	fetched, err := repo.GetAuthorityByAuthority(ctx, authority.Authority)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, authority.ID, fetched.ID)
	assert.Equal(t, authority.Authority, fetched.Authority)

	// Set the authority
	changed, err := repo.SetAuthority(ctx, authority.ID, &repository.AuthorityData{
		Authority: "ROLE_TEST_CHANGED",
	})
	assert.NoError(t, err)
	assert.NotNil(t, changed)
	assert.Equal(t, authority.ID, changed.ID)
	assert.Equal(t, "ROLE_TEST_CHANGED", changed.Authority)

	// Get the authority by id
	fetched, err = repo.GetAuthorityById(ctx, changed.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, changed.ID, fetched.ID)
	assert.Equal(t, changed.Authority, fetched.Authority)

	// Search authority
	_, err = repo.AddAuthority(ctx, &repository.AuthorityData{
		Authority: "ROLE_TEST_OTHER",
	})
	assert.NoError(t, err)

	page, err := repo.SearchAuthorities(ctx,
		&repository.SearchAuthoritiesCriteria{SearchField: changed.Authority},
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

	page, err = repo.SearchAuthorities(ctx,
		&repository.SearchAuthoritiesCriteria{},
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

	// Delete the authority
	err = repo.DeleteAuthorityById(ctx, authority.ID)
	assert.NoError(t, err)

	// Ensure it’s deleted
	_, err = repo.GetAuthorityById(ctx, authority.ID)
	assert.Error(t, err)
}
