package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
)

type userController struct {
	userService service.UserService
}

var _ openapi.UserControllerAPI = (*userController)(nil)

func NewUserController(userService service.UserService) openapi.UserControllerAPI {
	return &userController{userService}
}

func (u userController) AddUser(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) DeleteUser(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) GetUser(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) GetUsers(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) SetAuthorities(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) SetConfirmed(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) SetEnabled(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (u userController) SetUser(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
