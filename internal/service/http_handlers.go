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

var _ security.HttpHandlers[*openapi.UserDetail] = (*httpHandlers)(nil)

func NewHttpHandlers(jwtService *JwtService, userRepository repository.UserRepository) security.HttpHandlers[*openapi.UserDetail] {
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

func (h *httpHandlers) DecodeUserDetail(c *gin.Context, token string) (*openapi.UserDetail, error) {
	jwtToken, err := h.jwtService.GetAccessJwtToken(c.Request.Context())
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	id, _, err := h.jwtService.ParseAuthToken(c.Request.Context(), jwtToken, token)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	user, err := h.userRepository.GetUser(c.Request.Context(), id)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	return mapUserDetail(c.Request.Context(), h.userRepository, user)
}

func (h *httpHandlers) GetUserAuthorities(c *gin.Context, userDetail *openapi.UserDetail) ([]string, error) {
	var authorities = make([]string, len(userDetail.Authorities))
	for i, authority := range userDetail.Authorities {
		authorities[i] = authority.Authority
	}
	return authorities, nil
}
