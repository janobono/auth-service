package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"net/http"
)

type healthController struct {
}

var _ openapi.HealthControllerAPI = (*healthController)(nil)

func NewHealthController() openapi.HealthControllerAPI {
	return &healthController{}
}

func (h healthController) Livez(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, openapi.HealthStatus{Status: "UP"})
}

func (h healthController) Readyz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, openapi.HealthStatus{Status: "READY"})
}
