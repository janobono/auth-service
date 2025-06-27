package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
)

type authController struct {
	authService service.AuthService
}

var _ openapi.AuthControllerAPI = (*authController)(nil)

func NewAuthController(authService service.AuthService) openapi.AuthControllerAPI {
	return &authController{authService}
}

func (a *authController) ChangeEmail(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) ChangePassword(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) ChangeUserAttributes(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) Confirm(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) GetUserDetail(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) Refresh(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) ResetPassword(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) SignIn(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *authController) SignUp(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
