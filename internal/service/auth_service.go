package service

import (
	"context"
	"github.com/janobono/auth-service/api"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/db/repository"
	"github.com/janobono/auth-service/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"time"
)

type AuthService struct {
	api.UnimplementedAuthServiceServer
	dataSource *db.DataSource
	jwtToken   *util.JwtToken
}

func NewAuthService(dataSource *db.DataSource, jwtToken *util.JwtToken) *AuthService {
	return &AuthService{
		dataSource: dataSource,
		jwtToken:   jwtToken,
	}
}

func (as *AuthService) SignIn(ctx context.Context, signInData *api.SignInData) (*wrapperspb.StringValue, error) {
	email := util.ToScDf(signInData.Email)

	saUser, err := as.dataSource.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid credentials")
	}

	err = as.checkConfirmed(&saUser)
	if err != nil {
		return nil, err
	}

	err = as.checkEnabled(&saUser)
	if err != nil {
		return nil, err
	}

	err = as.checkPassword(&saUser, signInData)
	if err != nil {
		return nil, err
	}

	authorities, err := as.getAuthorities(ctx, saUser.ID)

	return as.createAuthToken(saUser.ID, authorities)
}

func (as *AuthService) SignUp(ctx context.Context, signUpData *api.SignUpData) (*wrapperspb.StringValue, error) {
	email := util.ToScDf(signUpData.Email)

	count, err := as.dataSource.Queries.CountUsersByCriteria(ctx, repository.CountUsersByCriteriaParams{
		Column2: email,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to check email")
	}
	if count > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Email is already used")
	}

	password, err := util.Encode(signUpData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to encrypt password")
	}

	saUser, err := as.dataSource.Queries.AddUser(ctx, repository.AddUserParams{
		Email:     email,
		Password:  password,
		FirstName: signUpData.FirstName,
		LastName:  signUpData.LastName,
		Confirmed: false,
		Enabled:   true,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Add user failed")
	}

	return as.createAuthToken(saUser.ID, &[]string{})
}

func (as *AuthService) checkConfirmed(user *repository.SaUser) error {
	if !user.Confirmed {
		return status.Errorf(codes.PermissionDenied, "Account not confirmed")
	}
	return nil
}

func (as *AuthService) checkEnabled(user *repository.SaUser) error {
	if !user.Enabled {
		return status.Errorf(codes.PermissionDenied, "Account not enabled")
	}
	return nil
}

func (as *AuthService) checkPassword(user *repository.SaUser, signInData *api.SignInData) error {
	if err := util.Compare(signInData.Password, user.Password); err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid credentials")
	}
	return nil
}

func (as *AuthService) createAuthToken(id int64, authorities *[]string) (*wrapperspb.StringValue, error) {
	token, err := as.jwtToken.GenerateToken(&util.JwtContent{ID: id, Authorities: *authorities}, time.Now().Unix())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token")
	}
	return &wrapperspb.StringValue{Value: token}, nil
}

func (as *AuthService) getAuthorities(ctx context.Context, id int64) (*[]string, error) {
	saAuthorities, err := as.dataSource.Queries.GetUserAuthorities(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user authorities")
	}

	authorities := make([]string, len(saAuthorities))
	for i, saAuthority := range saAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return &authorities, nil
}
