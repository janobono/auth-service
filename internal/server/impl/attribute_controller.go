package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
)

type attributeController struct {
	AttributeService service.AttributeService
}

var _ openapi.AttributeControllerAPI = (*attributeController)(nil)

func NewAttributeController(attributeService service.AttributeService) openapi.AttributeControllerAPI {
	return &attributeController{attributeService}
}

func (a *attributeController) AddAttribute(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *attributeController) DeleteAttribute(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *attributeController) GetAttribute(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *attributeController) GetAttributes(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (a *attributeController) SetAttribute(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
