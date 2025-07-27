package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/go-util/common"
	db2 "github.com/janobono/go-util/db"
	"github.com/janobono/go-util/security"
	"github.com/samborkent/uuidv7"
	"net/http"
	"strconv"
	"strings"
)

func parseId(ctx *gin.Context) (pgtype.UUID, bool) {
	id := ctx.Param("id")
	if common.IsBlank(id) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, "'id' must not be blank")
		return pgtype.UUID{}, false
	}
	if !uuidv7.IsValidString(id) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, "'id' must be uuid v7")
		return pgtype.UUID{}, false
	}
	result, err := db2.ParseUUID(id)
	if err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, err.Error())
		return pgtype.UUID{}, false
	}
	return result, true
}

func parsePageable(ctx *gin.Context, defaultSort string) *common.Pageable {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "20"))
	sort := ctx.DefaultQuery("sort", defaultSort)

	return &common.Pageable{
		Page: int32(page),
		Size: int32(size),
		Sort: sort,
	}
}

func parseStringSlice(ctx *gin.Context, key string) []string {
	value := ctx.Query(key)
	if common.IsBlank(value) {
		return []string{}
	}
	return strings.Split(value, ",")
}

func getAccessToken(ctx *gin.Context) (string, bool) {
	token, ok := security.GetHttpAccessToken(ctx)

	if !ok {
		AbortWithStatus(ctx, http.StatusUnauthorized)
	}

	return token, ok
}

func getUserDetail(ctx *gin.Context) (*openapi.UserDetail, bool) {
	principal, ok := security.GetHttpUserDetail[*openapi.UserDetail](ctx)

	if !ok {
		AbortWithStatus(ctx, http.StatusUnauthorized)
	}

	return principal, ok
}
