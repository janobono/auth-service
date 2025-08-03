package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"log/slog"
	"net/http"
)

type jwksController struct {
	jwksService *service.JwkService
}

var _ openapi.JwksControllerAPI = (*jwksController)(nil)

func NewJwksController(jwksService *service.JwkService) openapi.JwksControllerAPI {
	return &jwksController{jwksService}
}

func (j *jwksController) GetJwks(ctx *gin.Context) {
	jwks, err := j.jwksService.GetJwks(ctx.Request.Context())
	if err != nil {
		slog.Error("Failed to get jwks", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, jwks)
}
