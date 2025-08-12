package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/stretchr/testify/assert"
)

// helper to make unique emails
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func TestUserRepository_CountsAndGetters(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)

	email := uniqueEmail("count_get")
	u, err := userRepository.AddUser(ctx, &repository.UserData{
		Email:     email,
		Password:  "pw",
		Enabled:   true,
		Confirmed: false,
	})
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = userRepository.DeleteUserById(ctx, u.ID)
	})

	// CountByEmail
	cnt, err := userRepository.CountByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), cnt)

	// CountById
	cnt, err = userRepository.CountById(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), cnt)

	// CountByEmailAndNotId
	cnt, err = userRepository.CountByEmailAndNotId(ctx, email, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), cnt)

	// Getters
	byEmail, err := userRepository.GetUserByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, u.ID, byEmail.ID)

	byID, err := userRepository.GetUserById(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, email, byID.Email)
}

func TestUserRepository_Setters(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)

	email := uniqueEmail("setters")
	u, err := userRepository.AddUser(ctx, &repository.UserData{
		Email:     email,
		Password:  "old",
		Enabled:   false,
		Confirmed: false,
	})
	assert.NoError(t, err)
	t.Cleanup(func() { _ = userRepository.DeleteUserById(ctx, u.ID) })

	// email
	newEmail := uniqueEmail("setters_new")
	u, err = userRepository.SetUserEmail(ctx, u.ID, newEmail)
	assert.NoError(t, err)
	assert.Equal(t, newEmail, u.Email)

	// password
	u, err = userRepository.SetUserPassword(ctx, u.ID, "newpw")
	assert.NoError(t, err)
	assert.Equal(t, "newpw", u.Password)

	// enabled
	u, err = userRepository.SetUserEnabled(ctx, u.ID, true)
	assert.NoError(t, err)
	assert.True(t, u.Enabled)

	// confirmed
	u, err = userRepository.SetUserConfirmed(ctx, u.ID, true)
	assert.NoError(t, err)
	assert.True(t, u.Confirmed)
}

func TestUserRepository_Search_ByEmailFilter(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)

	// create three users that differ in email
	base := fmt.Sprintf("search_email_%d", time.Now().UnixNano())
	u1, _ := userRepository.AddUser(ctx, &repository.UserData{Email: base + "+one@example.com", Password: "x", Enabled: true, Confirmed: true})
	u2, _ := userRepository.AddUser(ctx, &repository.UserData{Email: base + "+two@example.com", Password: "x", Enabled: true, Confirmed: true})
	u3, _ := userRepository.AddUser(ctx, &repository.UserData{Email: "other+" + base + "@example.com", Password: "x", Enabled: true, Confirmed: true})

	t.Cleanup(func() {
		_ = userRepository.DeleteUserById(ctx, u1.ID)
		_ = userRepository.DeleteUserById(ctx, u2.ID)
		_ = userRepository.DeleteUserById(ctx, u3.ID)
	})

	criteria := &repository.SearchUsersCriteria{
		Email:       base, // repo uses LIKE with %ToScDf(email)%
		SearchField: "",   // not used here
	}
	pageable := &common.Pageable{Page: 0, Size: 10, Sort: "u.created_at desc"}

	page, err := userRepository.SearchUsers(ctx, criteria, pageable)
	assert.NoError(t, err)
	// Both +one and +two match, the "other+base" also matches because contains 'base'
	assert.GreaterOrEqual(t, page.TotalElements, int64(3))
	emails := make([]string, 0, len(page.Content))
	for _, u := range page.Content {
		emails = append(emails, u.Email)
	}
	assert.Contains(t, emails, u1.Email)
	assert.Contains(t, emails, u2.Email)
	assert.Contains(t, emails, u3.Email)
}

func TestUserRepository_Search_BySearchField_And_AttributeKeys(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)
	attributeRepository := repository.NewAttributeRepository(DataSource)

	// Ensure attribute exists
	attr, err := attributeRepository.AddAttribute(ctx, &repository.AttributeData{
		Key:      "nickname",
		Required: false,
		Hidden:   false,
	})
	assert.NoError(t, err)

	// Users
	emailPrefix := fmt.Sprintf("search_attr_%d", time.Now().UnixNano())
	u1, _ := userRepository.AddUser(ctx, &repository.UserData{Email: emailPrefix + "+a@example.com", Password: "x", Enabled: true, Confirmed: true})
	u2, _ := userRepository.AddUser(ctx, &repository.UserData{Email: emailPrefix + "+b@example.com", Password: "x", Enabled: true, Confirmed: true})
	t.Cleanup(func() {
		_ = userRepository.DeleteUserById(ctx, u1.ID)
		_ = userRepository.DeleteUserById(ctx, u2.ID)
	})

	// Attributes (so search can hit attribute value)
	_, err = userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID: u1.ID,
		Attributes: []*repository.UserAttribute{
			{Attribute: attr, Value: "iron man"},
		},
	})
	assert.NoError(t, err)

	_, err = userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID: u2.ID,
		Attributes: []*repository.UserAttribute{
			{Attribute: attr, Value: "captain"},
		},
	})
	assert.NoError(t, err)

	// Search value "iron" against attribute key "nickname"
	criteria := &repository.SearchUsersCriteria{
		SearchField:   "iron",               // repo splits by space and searches email + attribute values
		AttributeKeys: []string{"nickname"}, // restrict attribute joins
	}
	pageable := &common.Pageable{Page: 0, Size: 10, Sort: "u.created_at asc"}

	page, err := userRepository.SearchUsers(ctx, criteria, pageable)
	assert.NoError(t, err)

	// Should include u1, not u2
	emails := make([]string, 0, len(page.Content))
	for _, u := range page.Content {
		emails = append(emails, u.Email)
	}
	assert.Contains(t, emails, u1.Email)
	assert.NotContains(t, emails, u2.Email)
}

func TestUserRepository_AddUserWithAttributesAndAuthorities_Tx(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)
	authorityRepository := repository.NewAuthorityRepository(DataSource)
	attributeRepository := repository.NewAttributeRepository(DataSource)

	// Seed authority & attribute
	auth, err := authorityRepository.AddAuthority(ctx, &repository.AuthorityData{Authority: "ROLE_TX"})
	assert.NoError(t, err)
	attr, err := attributeRepository.AddAttribute(ctx, &repository.AttributeData{Key: "title", Required: false, Hidden: false})
	assert.NoError(t, err)

	email := uniqueEmail("tx")
	u, err := userRepository.AddUserWithAttributesAndAuthorities(ctx,
		&repository.UserData{
			Email:     email,
			Password:  "pw",
			Enabled:   true,
			Confirmed: true,
		},
		[]*repository.UserAttribute{{Attribute: attr, Value: "Dr."}},
		[]*repository.Authority{auth},
	)
	assert.NoError(t, err)
	assert.Equal(t, email, u.Email)
	t.Cleanup(func() { _ = userRepository.DeleteUserById(ctx, u.ID) })

	// Verify they were set
	gotAttrs, err := userRepository.GetUserAttributes(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, gotAttrs, 1)
	assert.Equal(t, "Dr.", gotAttrs[0].Value)

	gotAuths, err := userRepository.GetUserAuthorities(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, gotAuths, 1)
	assert.Equal(t, "ROLE_TX", gotAuths[0].Authority)
}

func TestUserRepository_SetUserAttributes_Replaces(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)
	attributeRepository := repository.NewAttributeRepository(DataSource)

	attr1, err := attributeRepository.AddAttribute(ctx, &repository.AttributeData{Key: "k1", Required: false, Hidden: false})
	assert.NoError(t, err)
	attr2, err := attributeRepository.AddAttribute(ctx, &repository.AttributeData{Key: "k2", Required: false, Hidden: false})
	assert.NoError(t, err)

	u, err := userRepository.AddUser(ctx, &repository.UserData{
		Email:     uniqueEmail("replace_attr"),
		Password:  "x",
		Enabled:   true,
		Confirmed: true,
	})
	assert.NoError(t, err)
	t.Cleanup(func() { _ = userRepository.DeleteUserById(ctx, u.ID) })

	_, err = userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID: u.ID,
		Attributes: []*repository.UserAttribute{
			{Attribute: attr1, Value: "v1"},
			{Attribute: attr2, Value: "v2"},
		},
	})
	assert.NoError(t, err)

	// Replace with only one attribute
	_, err = userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID: u.ID,
		Attributes: []*repository.UserAttribute{
			{Attribute: attr2, Value: "v2-new"},
		},
	})
	assert.NoError(t, err)

	got, err := userRepository.GetUserAttributes(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "k2", got[0].Attribute.Key)
	assert.Equal(t, "v2-new", got[0].Value)
}

func TestUserRepository_SetUserAuthorities_Replaces(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)
	authorityRepository := repository.NewAuthorityRepository(DataSource)

	a1, err := authorityRepository.AddAuthority(ctx, &repository.AuthorityData{Authority: "ROLE_A"})
	assert.NoError(t, err)
	a2, err := authorityRepository.AddAuthority(ctx, &repository.AuthorityData{Authority: "ROLE_B"})
	assert.NoError(t, err)

	u, err := userRepository.AddUser(ctx, &repository.UserData{
		Email:     uniqueEmail("replace_auth"),
		Password:  "x",
		Enabled:   true,
		Confirmed: true,
	})
	assert.NoError(t, err)
	t.Cleanup(func() { _ = userRepository.DeleteUserById(ctx, u.ID) })

	_, err = userRepository.SetUserAuthorities(ctx, &repository.UserAuthoritiesData{
		UserID:      u.ID,
		Authorities: []*repository.Authority{a1, a2},
	})
	assert.NoError(t, err)

	// Replace with single authority
	_, err = userRepository.SetUserAuthorities(ctx, &repository.UserAuthoritiesData{
		UserID:      u.ID,
		Authorities: []*repository.Authority{a2},
	})
	assert.NoError(t, err)

	got, err := userRepository.GetUserAuthorities(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "ROLE_B", got[0].Authority)
}

func TestUserRepository_DeleteUser_Cascades(t *testing.T) {
	ctx := context.Background()
	userRepository := repository.NewUserRepository(DataSource)
	authorityRepository := repository.NewAuthorityRepository(DataSource)
	attributeRepository := repository.NewAttributeRepository(DataSource)

	auth, _ := authorityRepository.AddAuthority(ctx, &repository.AuthorityData{Authority: "ROLE_DEL"})
	attr, _ := attributeRepository.AddAttribute(ctx, &repository.AttributeData{Key: "nickname", Required: false, Hidden: false})

	u, err := userRepository.AddUser(ctx, &repository.UserData{
		Email:     uniqueEmail("del"),
		Password:  "x",
		Enabled:   true,
		Confirmed: true,
	})
	assert.NoError(t, err)

	_, _ = userRepository.SetUserAuthorities(ctx, &repository.UserAuthoritiesData{
		UserID:      u.ID,
		Authorities: []*repository.Authority{auth},
	})
	_, _ = userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID: u.ID,
		Attributes: []*repository.UserAttribute{
			{Attribute: attr, Value: "bye"},
		},
	})

	// Act
	err = userRepository.DeleteUserById(ctx, u.ID)
	assert.NoError(t, err)

	// Assert user is gone
	cnt, err := userRepository.CountById(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), cnt)

	// And relations should be gone (ON DELETE CASCADE); expect empty results, not errors
	gotAttrs, err := userRepository.GetUserAttributes(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, gotAttrs, 0)

	gotAuths, err := userRepository.GetUserAuthorities(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, gotAuths, 0)
}
