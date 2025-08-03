package impl

import (
	"context"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/generated/proto"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/common"
	db2 "github.com/janobono/go-util/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log/slog"
)

type userServer struct {
	proto.UnimplementedUserServer
	userService *service.UserService
}

var _ proto.UserServer = (*userServer)(nil)

func NewUserServer(userService *service.UserService) proto.UserServer {
	return &userServer{userService: userService}
}

func (us *userServer) SearchUsers(ctx context.Context, searchCriteria *proto.SearchCriteria) (*proto.UserPage, error) {
	var pageable *common.Pageable

	if searchCriteria.Page == nil {
		pageable = &common.Pageable{
			Page: 0,
			Size: 20,
			Sort: "id asc",
		}
	} else {
		pageable = &common.Pageable{
			Page: searchCriteria.Page.Page,
			Size: searchCriteria.Page.Size,
			Sort: searchCriteria.Page.Sort,
		}
	}

	page, err := us.userService.GetUsers(ctx,
		&service.SearchUserCriteria{
			Email:         searchCriteria.Email,
			SearchField:   searchCriteria.SearchField,
			AttributeKeys: searchCriteria.AttributeKeys,
		},
		pageable,
	)

	if err != nil {
		slog.Error("SearchUsers failed", "error", err)
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var content []*proto.UserDetail
	for _, userDetail := range page.Content {
		content = append(content, us.protoUser(userDetail))
	}

	return &proto.UserPage{
		Page: &proto.PageDetail{
			Page:          page.Pageable.Page,
			Size:          page.Pageable.Size,
			Sort:          page.Pageable.Sort,
			First:         page.First,
			Last:          page.Last,
			Empty:         page.Empty,
			TotalPages:    page.TotalPages,
			TotalElements: page.TotalElements,
		},
		Content: content,
	}, nil
}

func (us *userServer) GetUser(ctx context.Context, id *wrapperspb.StringValue) (*proto.UserDetail, error) {
	userId, err := db2.ParseUUID(id.GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	userDetail, err := us.userService.GetUser(ctx, userId)
	if err != nil {
		slog.Error("Get user failed", "userID", id.GetValue(), "error", err)
		switch {
		case common.IsCode(err, string(openapi.NOT_FOUND)):
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}
	return us.protoUser(userDetail), err
}

func (us *userServer) protoUser(userDetail *openapi.UserDetail) *proto.UserDetail {
	var authorities = make([]string, len(userDetail.Authorities))
	for i, authority := range userDetail.Authorities {
		authorities[i] = authority.Authority
	}

	attributes := make(map[string]string)
	for _, attribute := range userDetail.Attributes {
		attributes[attribute.Key] = attribute.Value
	}

	return &proto.UserDetail{
		Id:          userDetail.Id,
		Email:       userDetail.Email,
		CreatedAt:   timestamppb.New(userDetail.CreatedAt),
		Confirmed:   userDetail.Confirmed,
		Enabled:     userDetail.Enabled,
		Authorities: authorities,
		Attributes:  attributes,
	}
}
