package impl

import (
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/go-util/common"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RespondWithError(ctx *gin.Context, statusCode int, code openapi.ErrorCode, message string) {
	slog.Error("Error occurred", "code", code, "error", message)
	ctx.JSON(statusCode, openapi.ErrorMessage{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

func RespondWithServiceError(ctx *gin.Context, err error) {
	switch {
	case common.IsCode(err, string(openapi.UNAUTHORIZED)):
		RespondWithError(ctx, http.StatusUnauthorized, openapi.UNAUTHORIZED, err.Error())
	case common.IsCode(err, string(openapi.FORBIDDEN)):
		RespondWithError(ctx, http.StatusForbidden, openapi.FORBIDDEN, err.Error())
	case common.IsCode(err, string(openapi.NOT_FOUND)):
		RespondWithError(ctx, http.StatusNotFound, openapi.NOT_FOUND, err.Error())
	case common.IsCode(err, string(openapi.INVALID_ARGUMENT)):
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, err.Error())
	default:
		RespondWithError(ctx, http.StatusInternalServerError, openapi.UNKNOWN, err.Error())
	}
}
