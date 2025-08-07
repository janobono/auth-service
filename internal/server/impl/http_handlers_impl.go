package impl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/security"
)

type httpHandlers struct {
	jwtService  *service.JwtService
	userService *service.UserService
}

var _ security.HttpHandlers[*openapi.UserDetail] = (*httpHandlers)(nil)

func NewHttpHandlers(jwtService *service.JwtService, userService *service.UserService) security.HttpHandlers[*openapi.UserDetail] {
	return &httpHandlers{jwtService, userService}
}

func (h *httpHandlers) MissingAuthorizationHeader(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

func (h *httpHandlers) Unauthorized(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

func (h *httpHandlers) PermissionDenied(c *gin.Context) {
	c.AbortWithStatus(http.StatusForbidden)
}

func (h *httpHandlers) DecodeUserDetail(c *gin.Context, token string) (*openapi.UserDetail, error) {
	jwtToken, err := h.jwtService.GetAccessJwtToken(c.Request.Context())
	if err != nil {
		return nil, err
	}

	id, _, err := h.jwtService.ParseAuthToken(c.Request.Context(), jwtToken, token)
	if err != nil {
		return nil, err
	}

	return h.userService.GetUser(c.Request.Context(), id)
}

func (h *httpHandlers) GetUserAuthorities(c *gin.Context, userDetail *openapi.UserDetail) ([]string, error) {
	var authorities = make([]string, len(userDetail.Authorities))
	for i, authority := range userDetail.Authorities {
		authorities[i] = authority.Authority
	}
	return authorities, nil
}
