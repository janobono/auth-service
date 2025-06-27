package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
)

type authorityController struct {
	authorityService service.AuthorityService
}

var _ openapi.AuthorityControllerAPI = (*authorityController)(nil)

func NewAuthorityController(authorityService service.AuthorityService) openapi.AuthorityControllerAPI {
	return &authorityController{authorityService}
}

func (a *authorityController) AddAuthority(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityController) DeleteAuthority(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityController) GetAuthorities(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityController) GetAuthority(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityController) SetAuthority(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
