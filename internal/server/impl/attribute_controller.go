package impl

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/common"
)

type attributeController struct {
	attributeService *service.AttributeService
}

var _ openapi.AttributeControllerAPI = (*attributeController)(nil)

func NewAttributeController(attributeService *service.AttributeService) openapi.AttributeControllerAPI {
	return &attributeController{attributeService}
}

func (a *attributeController) AddAttribute(ctx *gin.Context) {
	var data openapi.AttributeData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_BODY, "Invalid request body")
		return
	}
	if common.IsBlank(data.Key) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, "'key' must not be blank")
		return
	}

	attribute, err := a.attributeService.AddAttribute(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to add attribute", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, attribute)
}

func (a *attributeController) DeleteAttribute(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	err := a.attributeService.DeleteAttribute(ctx.Request.Context(), id)
	if err != nil {
		slog.Error("Failed to delete attribute", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (a *attributeController) GetAttribute(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	attribute, err := a.attributeService.GetAttribute(ctx.Request.Context(), id)
	if err != nil {
		slog.Error("Failed to get attribute", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, attribute)
}

func (a *attributeController) GetAttributes(ctx *gin.Context) {
	result, err := a.attributeService.GetAttributes(
		ctx.Request.Context(),
		&service.SearchAttributeCriteria{
			SearchField: ctx.Query("searchField"),
		},
		parsePageable(ctx, "key ASC"))

	if err != nil {
		slog.Error("Failed to get attributes", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (a *attributeController) SetAttribute(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	var data openapi.AttributeData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_BODY, "Invalid request body")
		return
	}
	if common.IsBlank(data.Key) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_FIELD, "'key' must not be blank")
		return
	}

	attribute, err := a.attributeService.SetAttribute(ctx.Request.Context(), id, &data)
	if err != nil {
		slog.Error("Failed to update attribute", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, attribute)
}
