package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/common"
	"log/slog"
	"net/http"
)

type authController struct {
	authService service.AuthService
}

var _ openapi.AuthControllerAPI = (*authController)(nil)

func NewAuthController(authService service.AuthService) openapi.AuthControllerAPI {
	return &authController{authService}
}

func (a *authController) ChangeEmail(ctx *gin.Context) {
	var data openapi.ChangeEmail
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
	if common.IsBlank(data.Password) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'password' must not be blank")
		return
	}
	if common.IsBlank(data.CaptchaText) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaText' must not be blank")
		return
	}
	if common.IsBlank(data.CaptchaToken) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaToken' must not be blank")
		return
	}

	userDetail, ok := getUserDetail(ctx)
	if !ok {
		return
	}

	authentificationResponse, err := a.authService.ChangeEmail(ctx.Request.Context(), userDetail, &data)
	if err != nil {
		slog.Error("Failed to change email", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authentificationResponse)
}

func (a *authController) ChangePassword(ctx *gin.Context) {
	var data openapi.ChangePassword
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}
	if common.IsBlank(data.OldPassword) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'oldPassword' must not be blank")
		return
	}
	if common.IsBlank(data.NewPassword) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'newPassword' must not be blank")
		return
	}
	if common.IsBlank(data.CaptchaText) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaText' must not be blank")
		return
	}
	if common.IsBlank(data.CaptchaToken) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaToken' must not be blank")
		return
	}

	userDetail, ok := getUserDetail(ctx)
	if !ok {
		return
	}

	authentificationResponse, err := a.authService.ChangePassword(ctx.Request.Context(), userDetail, &data)
	if err != nil {
		slog.Error("Failed to change password", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authentificationResponse)
}

func (a *authController) ChangeUserAttributes(ctx *gin.Context) {
	var data openapi.ChangeUserAttributes
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}
	if common.IsBlank(data.CaptchaText) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaText' must not be blank")
		return
	}
	if common.IsBlank(data.CaptchaToken) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaToken' must not be blank")
		return
	}

	userDetail, ok := getUserDetail(ctx)
	if !ok {
		return
	}

	authentificationResponse, err := a.authService.ChangeUserAttributes(ctx.Request.Context(), userDetail, &data)
	if err != nil {
		slog.Error("Failed to change user attributes", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authentificationResponse)
}

func (a *authController) Confirm(ctx *gin.Context) {
	var data openapi.Confirmation
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}
	if common.IsBlank(data.Token) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'token' must not be blank")
		return
	}

	authentificationResponse, err := a.authService.Confirm(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to confirm", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authentificationResponse)
}

func (a *authController) GetUserDetail(ctx *gin.Context) {
	userDetail, ok := getUserDetail(ctx)
	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, userDetail)
}

func (a *authController) Refresh(ctx *gin.Context) {
	var data openapi.Refresh
	if err := ctx.ShouldBindJSON(&data); err != nil {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "Invalid request body")
		return
	}
	if common.IsBlank(data.RefreshToken) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'refreshToken' must not be blank")
		return
	}

	authentificationResponse, err := a.authService.RefreshToken(ctx.Request.Context(), data.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authentificationResponse)
}

func (a *authController) ResetPassword(ctx *gin.Context) {
	var data openapi.ResetPassword
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
	if common.IsBlank(data.CaptchaText) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaText' must not be blank")
		return
	}
	if common.IsBlank(data.CaptchaToken) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'captchaToken' must not be blank")
		return
	}

	err := a.authService.ResetPassword(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to reset password", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (a *authController) SignIn(ctx *gin.Context) {
	var data openapi.SignIn
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
	if common.IsBlank(data.Password) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'password' must not be blank")
		return
	}

	authentificationResponse, err := a.authService.SignIn(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to sign in", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, authentificationResponse)
}

func (a *authController) SignUp(ctx *gin.Context) {
	var data openapi.SignUp
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
	if common.IsBlank(data.Password) {
		RespondWithError(ctx, http.StatusBadRequest, openapi.INVALID_ARGUMENT, "'password' must not be blank")
		return
	}

	authentificationResponse, err := a.authService.SignUp(ctx.Request.Context(), &data)
	if err != nil {
		slog.Error("Failed to sign up", "error", err)
		RespondWithServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, authentificationResponse)
}
