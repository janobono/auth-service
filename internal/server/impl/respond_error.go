package impl

import (
	"errors"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/go-util/common"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func AbortWithStatus(ctx *gin.Context, statusCode int) {
	slog.Error("Abort with status", "status", statusCode)
	ctx.AbortWithStatus(statusCode)
}

func RespondWithError(ctx *gin.Context, statusCode int, code openapi.ErrorCode, message string) {
	slog.Error("Error occurred", "code", code, "error", message)
	ctx.JSON(statusCode, openapi.ErrorMessage{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

func RespondWithServiceError(ctx *gin.Context, err error) {
	var serviceErr *common.ServiceError
	if errors.Is(err, serviceErr) {
		RespondWithError(ctx, serviceErr.Status, openapi.ErrorCode(serviceErr.Code), serviceErr.Error())
		return
	}
	RespondWithError(ctx, http.StatusInternalServerError, openapi.UNKNOWN, err.Error())
}
