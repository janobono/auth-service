package repository_test

import (
	"context"
	"testing"

	"github.com/janobono/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_FullFlow(t *testing.T) {
	authorityRepository := repository.NewAuthorityRepository(DataSource)
	attributeRepository := repository.NewAttributeRepository(DataSource)
	userRepository := repository.NewUserRepository(DataSource)

	ctx := context.Background()

	// Add authority and attribute needed for setting later
	auth, err := authorityRepository.AddAuthority(ctx, &repository.AuthorityData{Authority: "ROLE_TEST"})
	assert.NoError(t, err)

	attr, err := attributeRepository.AddAttribute(ctx, &repository.AttributeData{
		Key:      "nickname",
		Required: false,
		Hidden:   false,
	})
	assert.NoError(t, err)

	// Add user
	user, err := userRepository.AddUser(ctx, &repository.UserData{
		Email:     "user@example.com",
		Password:  "securepass",
		Enabled:   true,
		Confirmed: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)

	// Get user by email
	fetchedByEmail, err := userRepository.GetUserByEmail(ctx, "user@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, fetchedByEmail.ID)

	// Set authorities
	updatedAuths, err := userRepository.SetUserAuthorities(ctx, &repository.UserAuthoritiesData{
		UserID:      user.ID,
		Authorities: []*repository.Authority{auth},
	})
	assert.NoError(t, err)
	assert.Len(t, updatedAuths, 1)
	assert.Equal(t, "ROLE_TEST", updatedAuths[0].Authority)

	// Set attributes
	updatedAttrs, err := userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID: user.ID,
		Attributes: []*repository.UserAttribute{
			{
				Attribute: attr,
				Value:     "testnick",
			},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, updatedAttrs, 1)
	assert.Equal(t, "nickname", updatedAttrs[0].Attribute.Key)

	// Get authorities
	gotAuths, err := userRepository.GetUserAuthorities(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, gotAuths, 1)
	assert.Equal(t, "ROLE_TEST", gotAuths[0].Authority)

	// Get attributes
	gotAttrs, err := userRepository.GetUserAttributes(ctx, user.ID)
	assert.NoError(t, err)
	assert.Len(t, gotAttrs, 1)
	assert.Equal(t, "testnick", gotAttrs[0].Value)

	// Delete user
	err = userRepository.DeleteUserById(ctx, user.ID)
	assert.NoError(t, err)
}
