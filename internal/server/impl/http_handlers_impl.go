package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/security"
	"net/http"
	"time"
)

type httpHandlers struct {
	jwtService  *service.JwtService
	userService service.UserService
}

var _ security.HttpHandlers[*openapi.UserDetail] = (*httpHandlers)(nil)

func NewHttpHandlers(jwtService *service.JwtService, userService service.UserService) security.HttpHandlers[*openapi.UserDetail] {
	return &httpHandlers{jwtService, userService}
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
