package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/common"
	"log/slog"
	"net/http"
)

type authorityController struct {
	authorityService service.AuthorityService
}

var _ openapi.AuthorityControllerAPI = (*authorityController)(nil)

func NewAuthorityController(authorityService service.AuthorityService) openapi.AuthorityControllerAPI {
	return &authorityController{authorityService}
}

func (a *authorityController) AddAuthority(ctx *gin.Context) {
	var data openapi.AuthorityData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_BODY, "Invalid request body")
		return
	}
	if common.IsBlank(data.Authority) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, "'authority' must not be blank")
		return
	}

	authority, err := a.authorityService.AddAuthority(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to add authority", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, authority)
}

func (a *authorityController) DeleteAuthority(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	err := a.authorityService.DeleteAuthority(ctx.Request.Context(), id)
	if err != nil {
		slog.Error("Failed to delete authority", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (a *authorityController) GetAuthorities(ctx *gin.Context) {
	result, err := a.authorityService.GetAuthorities(
		ctx.Request.Context(),
		&service.SearchAuthorityCriteria{
			SearchField: ctx.Query("searchField"),
		},
		parsePageable(ctx, "authority ASC"))

	if err != nil {
		slog.Error("Failed to get authorities", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (a *authorityController) GetAuthority(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	authority, err := a.authorityService.GetAuthority(ctx.Request.Context(), id)
	if err != nil {
		slog.Error("Failed to get authority", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authority)
}

func (a *authorityController) SetAuthority(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	var data openapi.AuthorityData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_BODY, "Invalid request body")
		return
	}
	if common.IsBlank(data.Authority) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, "'authority' must not be blank")
		return
	}

	authority, err := a.authorityService.SetAuthority(ctx.Request.Context(), id, &data)
	if err != nil {
		slog.Error("Failed to update authority", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authority)
}
