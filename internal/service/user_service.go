package service

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/api"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/db/repository"
	"github.com/janobono/auth-service/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"strings"
)

type UserService struct {
	api.UnimplementedUserServiceServer
	dataSource *db.DataSource
}

func NewUserService(dataSource *db.DataSource) *UserService {
	return &UserService{dataSource: dataSource}
}

func (us *UserService) SearchUsers(ctx context.Context, searchCriteria *api.SearchCriteriaData) (*api.UsersPage, error) {
	countUsersCriteria := repository.CountUsersByCriteriaParams{}
	searchUsersCriteria := repository.SearchUsersByCriteriaParams{}
	countUsersCriteria.Column1 = util.ToScDf(searchCriteria.SearchField)
	searchUsersCriteria.Column1 = countUsersCriteria.Column1
	countUsersCriteria.Column2 = util.ToScDf(searchCriteria.Email)
	searchUsersCriteria.Column2 = countUsersCriteria.Column2

	count, err := us.dataSource.Queries.CountUsersByCriteria(ctx, countUsersCriteria)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Search users failed")
	}

	if count == 0 {
		return &api.UsersPage{
			TotalPages: 0,
			Page:       searchCriteria.Page,
			Size:       searchCriteria.Size,
			Content:    []*api.UserDetail{},
		}, nil
	}

	limit, offset := getLimitAndOffset(searchCriteria.Page, searchCriteria.Size)
	orderColumn, orderDirection := getOrderColumnAndDirection(searchCriteria.Sort)

	searchUsersCriteria.Limit = limit
	searchUsersCriteria.Offset = offset
	searchUsersCriteria.Column3 = fmt.Sprintf("%s %s", orderColumn, orderDirection)

	users, err := us.dataSource.Queries.SearchUsersByCriteria(ctx, searchUsersCriteria)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Search users failed")
	}

	content := make([]*api.UserDetail, len(users))
	for i, user := range users {
		authorities, err := us.getUserAuthorities(ctx, user)
		if err != nil {
			return nil, err
		}

		content[i] = &api.UserDetail{
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Confirmed:   user.Confirmed,
			Enabled:     user.Enabled,
			Authorities: *authorities,
		}
	}

	return &api.UsersPage{
		TotalPages: getTotalPages(count, int64(limit)),
		Page:       searchCriteria.Page,
		Size:       searchCriteria.Size,
		Content:    []*api.UserDetail{},
	}, nil
}

func (us *UserService) AddUser(ctx context.Context, userData *api.UserData) (*api.UserDetail, error) {
	email := util.ToScDf(userData.Email)
	password, err := util.Encode(userData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to encrypt password")
	}

	err = us.dataSource.ExecTx(ctx, func(queries *repository.Queries) error {
		user, err := queries.AddUser(ctx, repository.AddUserParams{
			Email:     email,
			Password:  password,
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			Confirmed: userData.Confirmed,
			Enabled:   userData.Enabled,
		})
		if err != nil {
			return err
		}

		err = queries.DeleteUserAuthorities(ctx, user.ID)
		if err != nil {
			return err
		}

		for _, authority := range userData.Authorities {
			saAuthority, err := queries.GetAuthority(ctx, authority)
			if err != nil {
				return err
			}

			err = queries.AddUserAuthority(ctx, repository.AddUserAuthorityParams{UserID: user.ID, AuthorityID: saAuthority.ID})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Add user failed")
	}

	return &api.UserDetail{
		Email:       email,
		FirstName:   userData.FirstName,
		LastName:    userData.LastName,
		Confirmed:   userData.Confirmed,
		Enabled:     userData.Enabled,
		Authorities: userData.Authorities,
	}, nil
}

func (us *UserService) GetUser(ctx context.Context, id *wrapperspb.Int64Value) (*api.UserDetail, error) {
	saUser, err := us.dataSource.Queries.GetUser(ctx, id.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get user failed")
	}

	authorities, err := us.getUserAuthorities(ctx, saUser)
	if err != nil {
		return nil, err
	}

	return &api.UserDetail{
		Email:       saUser.Email,
		FirstName:   saUser.FirstName,
		LastName:    saUser.LastName,
		Confirmed:   saUser.Confirmed,
		Enabled:     saUser.Enabled,
		Authorities: *authorities,
	}, nil
}

func (us *UserService) SetUser(ctx context.Context, userDataWithId *api.UserDataWithId) (*api.UserDetail, error) {
	email := util.ToScDf(userDataWithId.UserData.Email)

	saUser, err := us.dataSource.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to check email")
	}
	if saUser.ID != userDataWithId.Id {
		return nil, status.Errorf(codes.InvalidArgument, "Email is already used")
	}

	password, err := util.Encode(userDataWithId.UserData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to encrypt password")
	}

	err = us.dataSource.ExecTx(ctx, func(queries *repository.Queries) error {
		err := queries.SetUser(ctx, repository.SetUserParams{
			ID:        userDataWithId.Id,
			Email:     email,
			Password:  password,
			FirstName: userDataWithId.UserData.FirstName,
			LastName:  userDataWithId.UserData.LastName,
			Confirmed: userDataWithId.UserData.Confirmed,
			Enabled:   userDataWithId.UserData.Enabled,
		})
		if err != nil {
			return err
		}

		err = queries.DeleteUserAuthorities(ctx, userDataWithId.Id)
		if err != nil {
			return err
		}

		for _, authority := range userDataWithId.UserData.Authorities {
			saAuthority, err := queries.GetAuthority(ctx, authority)
			if err != nil {
				return err
			}

			err = queries.AddUserAuthority(ctx, repository.AddUserAuthorityParams{UserID: userDataWithId.Id, AuthorityID: saAuthority.ID})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Add user failed")
	}

	return &api.UserDetail{
		Email:       email,
		FirstName:   userDataWithId.UserData.FirstName,
		LastName:    userDataWithId.UserData.LastName,
		Confirmed:   userDataWithId.UserData.Confirmed,
		Enabled:     userDataWithId.UserData.Enabled,
		Authorities: userDataWithId.UserData.Authorities,
	}, nil
}

func (us *UserService) DeleteUser(ctx context.Context, id *wrapperspb.Int64Value) (*emptypb.Empty, error) {
	err := us.dataSource.Queries.DeleteUser(ctx, id.Value)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Delete user failed")
	}

	return &emptypb.Empty{}, nil
}

func getLimitAndOffset(page, size int32) (int32, int32) {
	if size <= 0 {
		size = 20
	}
	if page < 0 {
		page = 0
	}
	return size, page * size
}

func getOrderColumnAndDirection(sort string) (string, string) {
	sortArray := strings.Split(strings.ToLower(strings.TrimSpace(sort)), ",")
	if len(sortArray) == 2 {
		return sortArray[0], sortArray[1]
	}
	return "id", "asc"
}

func getTotalPages(count, limit int64) int64 {
	if count == 0 {
		return 0
	}
	if limit == 0 {
		return 0
	}
	result := count / limit
	if count%limit > 0 {
		result++
	}
	return result
}

func (us *UserService) getUserAuthorities(ctx context.Context, user repository.SaUser) (*[]string, error) {
	saAuthorities, err := us.dataSource.Queries.GetUserAuthorities(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get user authorities failed")
	}

	var authorities = make([]string, len(saAuthorities))
	for i, saAuthority := range saAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return &authorities, nil
}
