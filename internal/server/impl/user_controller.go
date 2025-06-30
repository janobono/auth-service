package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/common"
	"log/slog"
	"net/http"
)

type userController struct {
	userService service.UserService
}

var _ openapi.UserControllerAPI = (*userController)(nil)

func NewUserController(userService service.UserService) openapi.UserControllerAPI {
	return &userController{userService}
}

func (u userController) AddUser(ctx *gin.Context) {
	var data openapi.UserData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}
	if common.IsBlank(data.Email) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'email' must not be blank")
		return
	}
	if !common.IsValidEmail(data.Email) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'email' invalid format")
		return
	}

	user, err := u.userService.AddUser(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to add user", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (u userController) DeleteUser(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	err := u.userService.DeleteUser(ctx.Request.Context(), id)
	if err != nil {
		slog.Error("Failed to delete user", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (u userController) GetUser(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	user, err := u.userService.GetUser(ctx.Request.Context(), id)
	if err != nil {
		slog.Error("Failed to get user", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u userController) GetUsers(ctx *gin.Context) {
	result, err := u.userService.GetUsers(
		ctx.Request.Context(),
		&service.SearchUserCriteria{
			Email:         ctx.Query("email"),
			SearchField:   ctx.Query("searchField"),
			AttributeKeys: parseStringSlice(ctx, "attributeKeys"),
		},
		parsePageable(ctx, "key ASC"))

	if err != nil {
		slog.Error("Failed to get users", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (u userController) SetAuthorities(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}
	var data openapi.UserAuthoritiesData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}

	user, err := u.userService.SetAuthorities(ctx.Request.Context(), id, &data)
	if err != nil {
		slog.Error("Failed to update user authorities", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u userController) SetConfirmed(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}
	var data openapi.BooleanValue
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}

	user, err := u.userService.SetConfirmed(ctx.Request.Context(), id, &data)
	if err != nil {
		slog.Error("Failed to update user authorities", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u userController) SetEnabled(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}
	var data openapi.BooleanValue
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}

	user, err := u.userService.SetEnabled(ctx.Request.Context(), id, &data)
	if err != nil {
		slog.Error("Failed to update user authorities", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u userController) SetUser(ctx *gin.Context) {
	id, ok := parseId(ctx)
	if !ok {
		return
	}

	var data openapi.UserData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}
	if common.IsBlank(data.Email) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'email' must not be blank")
		return
	}

	user, err := u.userService.SetUser(ctx.Request.Context(), id, &data)
	if err != nil {
		slog.Error("Failed to update user", "id", id, "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}
