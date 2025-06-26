package service

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
	"net/http"
	"time"
)

type httpHandlers struct {
	jwtService     *JwtService
	userRepository repository.UserRepository
}

var _ security.HttpHandlers[*UserDetail] = (*httpHandlers)(nil)

func NewHttpHandlers(jwtService *JwtService, userRepository repository.UserRepository) security.HttpHandlers[*UserDetail] {
	return &httpHandlers{jwtService, userRepository}
}

func (h *httpHandlers) MissingAuthorizationHeader(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, openapi.ErrorMessage{
		Code:      openapi.UNAUTHORIZED,
		Message:   "Authorization header is missing.",
		Timestamp: time.Now().UTC(),
	})
}

func (h *httpHandlers) Unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, openapi.ErrorMessage{
		Code:      openapi.UNAUTHORIZED,
		Message:   "Invalid or missing authentication token.",
		Timestamp: time.Now().UTC(),
	})
}

func (h *httpHandlers) PermissionDenied(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, openapi.ErrorMessage{
		Code:      openapi.FORBIDDEN,
		Message:   "You are not authorized to access this resource.",
		Timestamp: time.Now().UTC(),
	})
}

func (h *httpHandlers) DecodeUserDetail(c *gin.Context, token string) (*UserDetail, error) {
	jwtToken, err := h.jwtService.GetAccessJwtToken(c)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	id, _, err := h.jwtService.ParseAuthToken(c, jwtToken, token)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	user, err := h.userRepository.GetUser(c, id)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	userAttributes, err := h.userRepository.GetUserAttributes(c, id)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	userAuthorities, err := h.userRepository.GetUserAuthorities(c, id)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	attributes := make([]*UserAttribute, 0, len(userAttributes))
	for _, userAttribute := range userAttributes {
		if !userAttribute.Attribute.Hidden {
			attributes = append(attributes, &UserAttribute{
				Attribute: &Attribute{
					Id:       userAttribute.Attribute.ID,
					Key:      userAttribute.Attribute.Key,
					Name:     userAttribute.Attribute.Name,
					Required: userAttribute.Attribute.Required,
					Hidden:   userAttribute.Attribute.Hidden,
				},
				Value: userAttribute.Value,
			})
		}
	}

	authorities := make([]*Authority, len(userAuthorities))
	for i, userAuthority := range userAuthorities {
		authorities[i] = &Authority{
			Id:        userAuthority.ID,
			Authority: userAuthority.Authority,
		}
	}

	return &UserDetail{
		Id:          user.ID,
		Email:       user.Email,
		Confirmed:   user.Confirmed,
		Enabled:     user.Enabled,
		Attributes:  attributes,
		Authorities: authorities,
	}, nil
}

func (h *httpHandlers) GetUserAuthorities(c *gin.Context, userDetail *UserDetail) ([]string, error) {
	var authorities = make([]string, len(userDetail.Authorities))
	for i, authority := range userDetail.Authorities {
		authorities[i] = authority.Authority
	}
	return authorities, nil
}
