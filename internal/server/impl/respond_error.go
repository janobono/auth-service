package impl

import (
	"github.com/janobono/auth-service/generated/openapi"
	"time"

	"github.com/gin-gonic/gin"
)

func RespondWithError(ctx *gin.Context, statusCode int, code openapi.ErrorCode, message string) {
	ctx.JSON(statusCode, openapi.ErrorMessage{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}
