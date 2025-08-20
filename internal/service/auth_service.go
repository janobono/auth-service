package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/generated/proto"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/auth-service/internal/service/client"
	"github.com/janobono/go-util/common"
	db2 "github.com/janobono/go-util/db"
	"github.com/janobono/go-util/security"
)

const (
	CONFIRMATION_TYPE = "CONFIRMATION_TYPE"
	ID                = "ID"
	PASSWORD          = "PASSWORD"
	CONFIRM_USER      = "CONFIRM_USER"
	RESET_PASSWORD    = "RESET_PASSWORD"
)

type AuthService struct {
	appConfig           *config.AppConfig
	mailConfig          *config.MailConfig
	passwordEncoder     *security.PasswordEncoder
	randomString        *security.RandomString
	captchaClient       client.CaptchaClient
	mailClient          client.MailClient
	jwtService          *JwtService
	attributeRepository repository.AttributeRepository
	authorityRepository repository.AuthorityRepository
	userRepository      repository.UserRepository
}

func NewAuthService(
	appConfig *config.AppConfig,
	mailConfig *config.MailConfig,
	passwordEncoder *security.PasswordEncoder,
	randomString *security.RandomString,
	captchaClient client.CaptchaClient,
	mailClient client.MailClient,
	jwtService *JwtService,
	attributeRepository repository.AttributeRepository,
	authorityRepository repository.AuthorityRepository,
	userRepository repository.UserRepository,
) *AuthService {
	return &AuthService{
		appConfig:           appConfig,
		mailConfig:          mailConfig,
		passwordEncoder:     passwordEncoder,
		randomString:        randomString,
		captchaClient:       captchaClient,
		mailClient:          mailClient,
		jwtService:          jwtService,
		attributeRepository: attributeRepository,
		authorityRepository: authorityRepository,
		userRepository:      userRepository,
	}
}

func (as *AuthService) ChangeEmail(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeEmail) (*openapi.AuthenticationResponse, error) {
	if err := as.checkCaptcha(ctx, data.CaptchaText, data.CaptchaToken); err != nil {
		return nil, err
	}

	user, err := as.getUser(ctx, userDetail.Id)
	if err != nil {
		return nil, err
	}

	count, err := as.userRepository.CountByEmailAndNotId(ctx, data.Email, user.ID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.EMAIL_ALREADY_EXISTS), "'email' already exists")
	}

	if err := as.checkPassword(user, data.Password); err != nil {
		return nil, err
	}

	user, err = as.userRepository.SetUserEmail(ctx, user.ID, data.Email)
	if err != nil {
		return nil, err
	}

	authorities, err := as.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) ChangePassword(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangePassword) (*openapi.AuthenticationResponse, error) {
	if err := as.checkCaptcha(ctx, data.CaptchaText, data.CaptchaToken); err != nil {
		return nil, err
	}

	user, err := as.getUser(ctx, userDetail.Id)
	if err != nil {
		return nil, err
	}

	if err := as.checkPassword(user, data.OldPassword); err != nil {
		return nil, err
	}

	password, err := as.passwordEncoder.Encode(data.NewPassword)
	if err != nil {
		return nil, err
	}

	user, err = as.userRepository.SetUserPassword(ctx, user.ID, password)
	if err != nil {
		return nil, err
	}

	authorities, err := as.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) ChangeUserAttributes(
	ctx context.Context,
	userDetail *openapi.UserDetail,
	data *openapi.ChangeUserAttributes,
) (*openapi.AuthenticationResponse, error) {
	if err := as.checkCaptcha(ctx, data.CaptchaText, data.CaptchaToken); err != nil {
		return nil, err
	}

	user, err := as.getUser(ctx, userDetail.Id)
	if err != nil {
		return nil, err
	}

	savedAttributes, err := as.userRepository.GetUserAttributes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	mandatoryAttributes := make(map[string]string, len(savedAttributes))
	for _, savedAttribute := range savedAttributes {
		mandatoryAttributes[savedAttribute.Attribute.Key] = savedAttribute.Value
	}

	userAttributes, err := as.createAttributes(ctx, data.Attributes, mandatoryAttributes)
	if err != nil {
		return nil, err
	}

	if _, err = as.userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID:     user.ID,
		Attributes: userAttributes,
	}); err != nil {
		return nil, err
	}

	authorities, err := as.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) Confirm(ctx context.Context, data *openapi.Confirmation) (*openapi.AuthenticationResponse, error) {
	confirmationData, err := as.parseConfirmationToken(ctx, data.Token)
	if err != nil {
		return nil, err
	}

	confirmationType, ok := confirmationData[CONFIRMATION_TYPE]
	if !ok {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_FIELD), "confirmation type not found")
	}

	var (
		user *repository.User
	)

	switch confirmationType {
	case CONFIRM_USER:
		user, err = as.confirmUser(ctx, confirmationData)
	case RESET_PASSWORD:
		user, err = as.resetPassword(ctx, confirmationData)
	default:
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_FIELD), "unsupported confirmation type")
	}
	if err != nil {
		return nil, err
	}

	authorities, err := as.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*openapi.AuthenticationResponse, error) {
	refreshJwt, err := as.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	id, authorities, err := as.jwtService.ParseAuthToken(ctx, refreshJwt, refreshToken)
	if err != nil {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_FIELD), err.Error())
	}

	accessJwt, err := as.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := as.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	return &openapi.AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) ResendConfirmation(ctx context.Context, data *openapi.ResendConfirmation) error {
	if err := as.checkCaptcha(ctx, data.CaptchaText, data.CaptchaToken); err != nil {
		return err
	}

	email := common.ToScDf(data.Email)

	user, err := as.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return as.sendConfirmationMail(ctx, user)
}

func (as *AuthService) ResetPassword(ctx context.Context, data *openapi.ResetPassword) error {
	if err := as.checkCaptcha(ctx, data.CaptchaText, data.CaptchaToken); err != nil {
		return err
	}

	email := common.ToScDf(data.Email)

	user, err := as.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "user not found")
	}

	if err := as.checkEnabled(user); err != nil {
		return err
	}

	return as.sendResetPasswordMail(ctx, user)
}

func (as *AuthService) SignIn(ctx context.Context, data *openapi.SignIn) (*openapi.AuthenticationResponse, error) {
	email := common.ToScDf(data.Email)

	user, err := as.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "user not found")
	}

	if err := as.checkEnabled(user); err != nil {
		return nil, err
	}

	if err := as.checkPassword(user, data.Password); err != nil {
		return nil, err
	}

	authorities, sErr := as.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, sErr
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) SignUp(ctx context.Context, data *openapi.SignUp) (*openapi.AuthenticationResponse, error) {
	if err := as.checkCaptcha(ctx, data.CaptchaText, data.CaptchaToken); err != nil {
		return nil, err
	}

	email := common.ToScDf(data.Email)
	password, err := as.passwordEncoder.Encode(data.Password)
	if err != nil {
		return nil, err
	}

	count, err := as.userRepository.CountByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.EMAIL_ALREADY_EXISTS), "'email' already exists")
	}

	userAttributes, err := as.createAttributes(ctx, data.Attributes, as.appConfig.MandatoryUserAttributes)
	if err != nil {
		return nil, err
	}

	userAuthorities, err := as.createAuthorities(ctx, as.appConfig.MandatoryUserAuthorities)
	if err != nil {
		return nil, err
	}

	user, err := as.userRepository.AddUserWithAttributesAndAuthorities(ctx, &repository.UserData{
		Email:     email,
		Password:  password,
		Confirmed: false,
		Enabled:   true,
	}, userAttributes, userAuthorities)
	if err != nil {
		return nil, err
	}

	authorities := make([]string, len(userAuthorities))
	for i, saAuthority := range userAuthorities {
		authorities[i] = saAuthority.Authority
	}

	err = as.sendConfirmationMail(ctx, user)
	if err != nil {
		return nil, err
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) checkCaptcha(ctx context.Context, captchaText string, captchaToken string) error {
	result, err := as.captchaClient.Validate(ctx, &proto.CaptchaData{
		Text:  captchaText,
		Token: captchaToken,
	})
	if err != nil {
		return err
	}

	if !result.Value {
		return common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_CAPTCHA), "invalid captcha")
	}

	return nil
}

func (as *AuthService) checkEnabled(user *repository.User) error {
	if !user.Enabled {
		return common.NewServiceError(http.StatusForbidden, string(openapi.USER_NOT_ENABLED), "account not enabled")
	}
	return nil
}

func (as *AuthService) checkPassword(user *repository.User, password string) error {
	if err := as.passwordEncoder.Compare(password, user.Password); err != nil {
		return common.NewServiceError(http.StatusForbidden, string(openapi.INVALID_CREDENTIALS), "wrong password")
	}
	return nil
}

func (as *AuthService) createAuthenticationResponse(ctx context.Context, id pgtype.UUID, authorities []string) (*openapi.AuthenticationResponse, error) {
	accessJwt, err := as.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := as.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	refreshJwt, err := as.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	refreshToken, err := as.jwtService.GenerateAuthToken(refreshJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	return &openapi.AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) createAttributes(
	ctx context.Context,
	attributes []openapi.AttributeValueData,
	mandatoryAttributes map[string]string,
) ([]*repository.UserAttribute, error) {

	userAttributeMap := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		userAttributeMap[attribute.Key] = attribute.Value
	}

	userAttributes := make([]*repository.UserAttribute, 0, len(userAttributeMap)+len(mandatoryAttributes))

	allAttributesMap, err := as.getAllAttributesMap(ctx)
	if err != nil {
		return nil, err
	}

	for k, v := range userAttributeMap {
		attr, ok := allAttributesMap[k]
		if !ok {
			return nil, common.NewServiceError(
				http.StatusBadRequest, string(openapi.INVALID_FIELD),
				fmt.Sprintf("unknown attribute key '%s'", k),
			)
		}
		if attr.Hidden {
			return nil, common.NewServiceError(
				http.StatusBadRequest, string(openapi.INVALID_FIELD),
				fmt.Sprintf("attribute '%s' is not settable", k),
			)
		}
		if common.IsBlank(v) {
			return nil, common.NewServiceError(
				http.StatusBadRequest, string(openapi.INVALID_FIELD),
				fmt.Sprintf("attribute '%s' value must not be blank", k),
			)
		}
		userAttributes = append(userAttributes, &repository.UserAttribute{Attribute: attr, Value: v})
	}

	for k, v := range mandatoryAttributes {
		attr, ok := allAttributesMap[k]
		if !ok {
			return nil, common.NewServiceError(
				http.StatusInternalServerError, string(openapi.INVALID_FIELD),
				fmt.Sprintf("mandatory attribute '%s' is not defined", k),
			)
		}
		if _, has := userAttributeMap[k]; !has {
			if common.IsBlank(v) {
				return nil, common.NewServiceError(
					http.StatusBadRequest, string(openapi.INVALID_FIELD),
					fmt.Sprintf("mandatory attribute '%s' value must not be blank", k),
				)
			}
			userAttributes = append(userAttributes, &repository.UserAttribute{Attribute: attr, Value: v})
		}
	}

	for k, attr := range allAttributesMap {
		if attr.Required && !attr.Hidden {
			if _, inUser := userAttributeMap[k]; !inUser {
				if _, inMandatory := mandatoryAttributes[k]; !inMandatory {
					return nil, common.NewServiceError(
						http.StatusBadRequest, string(openapi.REQUIRED_ATTRIBUTE),
						fmt.Sprintf("attribute '%s' is required", k),
					)
				}
			}
		}
	}

	return userAttributes, nil
}

func (as *AuthService) createAuthorities(
	ctx context.Context,
	mandatoryAuthorities []string,
) ([]*repository.Authority, error) {
	allAuthoritiesMap, err := as.getAllAuthoritiesMap(ctx)
	if err != nil {
		return nil, err
	}

	userAuthorities := make([]*repository.Authority, 0, len(mandatoryAuthorities))

	for _, k := range mandatoryAuthorities {
		if common.IsBlank(k) {
			return nil, common.NewServiceError(
				http.StatusBadRequest, string(openapi.INVALID_FIELD),
				"mandatory authority must not be blank",
			)
		}

		authority, ok := allAuthoritiesMap[k]
		if !ok {
			return nil, common.NewServiceError(
				http.StatusInternalServerError, string(openapi.INVALID_FIELD),
				fmt.Sprintf("mandatory authority '%s' is not defined", k),
			)
		}
		userAuthorities = append(userAuthorities, authority)
	}

	return userAuthorities, nil
}

func (as *AuthService) getAllAttributesMap(ctx context.Context) (map[string]*repository.Attribute, error) {
	allAttributes, err := as.attributeRepository.GetAllAttributes(ctx)
	if err != nil {
		return nil, err
	}

	allAttributesMap := make(map[string]*repository.Attribute, len(allAttributes))
	for _, attribute := range allAttributes {
		allAttributesMap[attribute.Key] = attribute
	}

	return allAttributesMap, nil
}

func (as *AuthService) getAllAuthoritiesMap(ctx context.Context) (map[string]*repository.Authority, error) {
	allAuthorities, err := as.authorityRepository.GetAllAuthorities(ctx)
	if err != nil {
		return nil, err
	}

	allAuthoritiesMap := make(map[string]*repository.Authority, len(allAuthorities))
	for _, authority := range allAuthorities {
		allAuthoritiesMap[authority.Authority] = authority
	}

	return allAuthoritiesMap, nil
}

func (as *AuthService) getAuthorities(ctx context.Context, id pgtype.UUID) ([]string, error) {
	userAuthorities, err := as.userRepository.GetUserAuthorities(ctx, id)
	if err != nil {
		return nil, err
	}

	authorities := make([]string, len(userAuthorities))
	for i, saAuthority := range userAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return authorities, nil
}

func (as *AuthService) getUser(ctx context.Context, id string) (*repository.User, error) {
	userId, err := db2.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	user, err := as.userRepository.GetUserById(ctx, userId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "User not found")
	}

	if err := as.checkEnabled(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (as *AuthService) generateConfirmationToken(ctx context.Context, data map[string]string) (string, error) {
	jwtToken, err := as.jwtService.GetConfirmJwtToken(ctx)
	if err != nil {
		return "", err
	}

	claims := make(jwt.MapClaims, len(data))
	for k, v := range claims {
		claims[k] = v.(string)
	}

	token, err := jwtToken.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (as *AuthService) parseConfirmationToken(ctx context.Context, token string) (map[string]string, error) {
	jwtToken, err := as.jwtService.GetConfirmJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := jwtToken.ParseToken(ctx, token)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(claims))

	for k, v := range claims {
		out[k] = v.(string)
	}

	return out, nil
}

func (as *AuthService) confirmUser(ctx context.Context, confirmData map[string]string) (*repository.User, error) {
	tokenId, ok := confirmData[ID]
	if !ok {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_TOKEN), "invalid token")
	}

	user, err := as.getUser(ctx, tokenId)
	if err != nil {
		return nil, err
	}

	user, err = as.userRepository.SetUserConfirmed(ctx, user.ID, true)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (as *AuthService) resetPassword(ctx context.Context, confirmData map[string]string) (*repository.User, error) {
	tokenId, ok := confirmData[ID]
	if !ok {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_TOKEN), "invalid token")
	}

	password, ok := confirmData[PASSWORD]
	if !ok {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_TOKEN), "invalid token")
	}

	user, err := as.getUser(ctx, tokenId)
	if err != nil {
		return nil, err
	}

	user, err = as.userRepository.SetUserPassword(ctx, user.ID, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (as *AuthService) tokenURL(token string) string {
	encodedToken := url.QueryEscape(token)
	return as.appConfig.ConfirmationWebUrl + as.appConfig.ConfirmationPath + encodedToken
}

func (as *AuthService) formatBody(templateUrl string, data interface{}) (string, error) {
	mailTemplateString, err := os.ReadFile(templateUrl)
	if err != nil {
		return "", err
	}

	mailTemplate, err := template.New("mail").Parse(string(mailTemplateString))
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := mailTemplate.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (as *AuthService) sendConfirmationMail(ctx context.Context, user *repository.User) error {
	if !as.appConfig.SignUpConfirmationMailEnabled {
		return nil
	}

	tokenData := make(map[string]string)
	tokenData[ID] = user.ID.String()

	token, err := as.generateConfirmationToken(ctx, tokenData)
	if err != nil {
		return err
	}

	confirmationUrl := as.tokenURL(token)

	body, err := as.formatBody(as.mailConfig.SignUpMailTemplateUrl, struct {
		ConfirmationUrl string
	}{
		ConfirmationUrl: confirmationUrl,
	})
	if err != nil {
		return err
	}

	mailData := client.NewMailData()

	mailData.From = as.mailConfig.User
	mailData.Recipients = []string{user.Email}
	mailData.Subject = as.mailConfig.SignUpMailSubject
	mailData.Body = body

	as.mailClient.SendEmail(mailData)
	return nil
}

func (as *AuthService) sendResetPasswordMail(ctx context.Context, user *repository.User) error {
	newPassword, err := as.randomString.Generate()
	if err != nil {
		return err
	}

	tokenData := make(map[string]string)
	tokenData[ID] = user.ID.String()
	tokenData[PASSWORD] = newPassword

	token, err := as.generateConfirmationToken(ctx, tokenData)
	if err != nil {
		return err
	}

	confirmationUrl := as.tokenURL(token)

	body, err := as.formatBody(as.mailConfig.ResetPasswordMailTemplateUrl, struct {
		NewPassword     string
		ConfirmationUrl string
	}{
		NewPassword:     newPassword,
		ConfirmationUrl: confirmationUrl,
	})
	if err != nil {
		return err
	}

	mailData := client.NewMailData()

	mailData.From = as.mailConfig.User
	mailData.Recipients = []string{user.Email}
	mailData.Subject = as.mailConfig.ResetPasswordMailSubject
	mailData.Body = body

	as.mailClient.SendEmail(mailData)
	return nil
}
