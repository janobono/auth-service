package impl

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/go-util/security"
)

type RouterContext struct {
	HandleFunctions  openapi.ApiHandleFunctions
	ContextPath      string
	ReadAuthorities  []string
	WriteAuthorities []string
	HttpHandlers     security.HttpHandlers[*openapi.UserDetail]
}

func NewRouter(routerContext RouterContext) *gin.Engine {
	router := gin.Default()

	authMiddleware := security.NewHttpTokenMiddleware[*openapi.UserDetail](security.HttpSecurityConfig{
		PublicEndpoints: map[string]struct{}{
			fmt.Sprintf("POST:%s/confirm", routerContext.ContextPath):              {},
			fmt.Sprintf("POST:%s/resend-confirmation", routerContext.ContextPath):  {},
			fmt.Sprintf("POST:%s/reset-password", routerContext.ContextPath):       {},
			fmt.Sprintf("POST:%s/sign-in", routerContext.ContextPath):              {},
			fmt.Sprintf("POST:%s/sign-up", routerContext.ContextPath):              {},
			fmt.Sprintf("GET:%s/livez", routerContext.ContextPath):                 {},
			fmt.Sprintf("GET:%s/readyz", routerContext.ContextPath):                {},
			fmt.Sprintf("GET:%s/.well-known/jwks.json", routerContext.ContextPath): {},
		},
		Authorities: map[string][]string{
			"GET:/attributes":     append(routerContext.ReadAuthorities, routerContext.WriteAuthorities...),
			"GET:/attributes/:id": append(routerContext.ReadAuthorities, routerContext.WriteAuthorities...),
			"POST:/attributes":    routerContext.WriteAuthorities,
			"PUT:/attributes/:id": routerContext.WriteAuthorities,

			"POST:/auth/change-email":           {},
			"POST:/auth/change-password":        {},
			"POST:/auth/change-user-attributes": {},
			"GET:/auth/user-detail":             {},

			"GET:/authorities":     append(routerContext.ReadAuthorities, routerContext.WriteAuthorities...),
			"GET:/authorities/:id": append(routerContext.ReadAuthorities, routerContext.WriteAuthorities...),
			"POST:/authorities":    routerContext.WriteAuthorities,
			"PUT:/authorities/:id": routerContext.WriteAuthorities,

			"GET:/users":                   append(routerContext.ReadAuthorities, routerContext.WriteAuthorities...),
			"GET:/users/:id":               append(routerContext.ReadAuthorities, routerContext.WriteAuthorities...),
			"DELETE:/users/:id":            routerContext.WriteAuthorities,
			"POST:/users":                  routerContext.WriteAuthorities,
			"PATCH:/users/:id/attributes":  routerContext.WriteAuthorities,
			"PATCH:/users/:id/authorities": routerContext.WriteAuthorities,
			"PATCH:/users/:id/confirm":     routerContext.WriteAuthorities,
			"PATCH:/users/:id/email":       routerContext.WriteAuthorities,
			"PATCH:/users/:id/enable":      routerContext.WriteAuthorities,
		},
	}, routerContext.HttpHandlers)

	// ✅ Apply to group
	group := router.Group(routerContext.ContextPath)
	group.Use(authMiddleware.HandlerFunc())

	// ✅ Register routes as usual
	for _, route := range getRoutes(routerContext.HandleFunctions) {
		handler := route.HandlerFunc
		if handler == nil {
			handler = openapi.DefaultHandleFunc
		}
		switch route.Method {
		case http.MethodGet:
			group.GET(route.Pattern, handler)
		case http.MethodPost:
			group.POST(route.Pattern, handler)
		case http.MethodPut:
			group.PUT(route.Pattern, handler)
		case http.MethodPatch:
			group.PATCH(route.Pattern, handler)
		case http.MethodDelete:
			group.DELETE(route.Pattern, handler)
		}
	}

	return router
}

func getRoutes(handleFunctions openapi.ApiHandleFunctions) []openapi.Route {
	return []openapi.Route{
		{
			"AddAttribute",
			http.MethodPost,
			"/attributes",
			handleFunctions.AttributeControllerAPI.AddAttribute,
		},
		{
			"DeleteAttribute",
			http.MethodDelete,
			"/attributes/:id",
			handleFunctions.AttributeControllerAPI.DeleteAttribute,
		},
		{
			"GetAttribute",
			http.MethodGet,
			"/attributes/:id",
			handleFunctions.AttributeControllerAPI.GetAttribute,
		},
		{
			"GetAttributes",
			http.MethodGet,
			"/attributes",
			handleFunctions.AttributeControllerAPI.GetAttributes,
		},
		{
			"SetAttribute",
			http.MethodPut,
			"/attributes/:id",
			handleFunctions.AttributeControllerAPI.SetAttribute,
		},
		{
			"ChangeEmail",
			http.MethodPost,
			"/auth/change-email",
			handleFunctions.AuthControllerAPI.ChangeEmail,
		},
		{
			"ChangePassword",
			http.MethodPost,
			"/auth/change-password",
			handleFunctions.AuthControllerAPI.ChangePassword,
		},
		{
			"ChangeUserAttributes",
			http.MethodPost,
			"/auth/change-user-attributes",
			handleFunctions.AuthControllerAPI.ChangeUserAttributes,
		},
		{
			"Confirm",
			http.MethodPost,
			"/auth/confirm",
			handleFunctions.AuthControllerAPI.Confirm,
		},
		{
			"GetUserDetail",
			http.MethodGet,
			"/auth/user-detail",
			handleFunctions.AuthControllerAPI.GetUserDetail,
		},
		{
			"Refresh",
			http.MethodPost,
			"/auth/refresh",
			handleFunctions.AuthControllerAPI.Refresh,
		},
		{
			"ResetConfirmation",
			http.MethodPost,
			"/auth/resend-confirmation",
			handleFunctions.AuthControllerAPI.ResendConfirmation,
		},
		{
			"ResetPassword",
			http.MethodPost,
			"/auth/reset-password",
			handleFunctions.AuthControllerAPI.ResetPassword,
		},
		{
			"SignIn",
			http.MethodPost,
			"/auth/sign-in",
			handleFunctions.AuthControllerAPI.SignIn,
		},
		{
			"SignUp",
			http.MethodPost,
			"/auth/sign-up",
			handleFunctions.AuthControllerAPI.SignUp,
		},
		{
			"AddAuthority",
			http.MethodPost,
			"/authorities",
			handleFunctions.AuthorityControllerAPI.AddAuthority,
		},
		{
			"DeleteAuthority",
			http.MethodDelete,
			"/authorities/:id",
			handleFunctions.AuthorityControllerAPI.DeleteAuthority,
		},
		{
			"GetAuthorities",
			http.MethodGet,
			"/authorities",
			handleFunctions.AuthorityControllerAPI.GetAuthorities,
		},
		{
			"GetAuthority",
			http.MethodGet,
			"/authorities/:id",
			handleFunctions.AuthorityControllerAPI.GetAuthority,
		},
		{
			"SetAuthority",
			http.MethodPut,
			"/authorities/:id",
			handleFunctions.AuthorityControllerAPI.SetAuthority,
		},
		{
			"Livez",
			http.MethodGet,
			"/livez",
			handleFunctions.HealthControllerAPI.Livez,
		},
		{
			"Readyz",
			http.MethodGet,
			"/readyz",
			handleFunctions.HealthControllerAPI.Readyz,
		},
		{
			"GetJwks",
			http.MethodGet,
			"/.well-known/jwks.json",
			handleFunctions.JwksControllerAPI.GetJwks,
		},
		{
			"AddUser",
			http.MethodPost,
			"/users",
			handleFunctions.UserControllerAPI.AddUser,
		},
		{
			"DeleteUser",
			http.MethodDelete,
			"/users/:id",
			handleFunctions.UserControllerAPI.DeleteUser,
		},
		{
			"GetUser",
			http.MethodGet,
			"/users/:id",
			handleFunctions.UserControllerAPI.GetUser,
		},
		{
			"GetUsers",
			http.MethodGet,
			"/users",
			handleFunctions.UserControllerAPI.GetUsers,
		},
		{
			"SetAttributes",
			http.MethodPatch,
			"/users/:id/attributes",
			handleFunctions.UserControllerAPI.SetAttributes,
		},
		{
			"SetAuthorities",
			http.MethodPatch,
			"/users/:id/authorities",
			handleFunctions.UserControllerAPI.SetAuthorities,
		},
		{
			"SetConfirmed",
			http.MethodPatch,
			"/users/:id/confirm",
			handleFunctions.UserControllerAPI.SetConfirmed,
		},
		{
			"SetEmail",
			http.MethodPatch,
			"/users/:id/email",
			handleFunctions.UserControllerAPI.SetEmail,
		},
		{
			"SetEnabled",
			http.MethodPatch,
			"/users/:id/enable",
			handleFunctions.UserControllerAPI.SetEnabled,
		},
	}
}
