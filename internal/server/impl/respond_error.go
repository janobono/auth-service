package impl

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/go-util/common"

	"github.com/gin-gonic/gin"
)

func AbortWithStatus(ctx *gin.Context, statusCode int) {
	if statusCode >= 500 {
		slog.Error("Abort with status", "status", statusCode)
	} else {
		slog.Warn("Abort with status", "status", statusCode)
	}
	ctx.AbortWithStatus(statusCode)
}

func RespondWithError(ctx *gin.Context, statusCode int, code openapi.ErrorCode, message string) {
	if statusCode >= 500 {
		slog.Error("Server error", "status", statusCode, "error_code", code, "message", message)
	} else {
		slog.Warn("Client error", "status", statusCode, "error_code", code, "message", message)
	}

	ctx.AbortWithStatusJSON(statusCode, openapi.ErrorMessage{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

func RespondWithServiceError(ctx *gin.Context, err error) {
	var serviceError *common.ServiceError
	if errors.As(err, &serviceError) {
		RespondWithError(ctx, serviceError.Status, openapi.ErrorCode(serviceError.Code), serviceError.Error())
		return
	}

	slog.Error("Unhandled error", "error", err)
	RespondWithError(ctx, http.StatusInternalServerError, openapi.UNKNOWN, "unexpected server error")
}
