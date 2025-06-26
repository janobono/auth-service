package impl

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"log/slog"
	"math/big"
	"net/http"
)

type jwksController struct {
	JwksService service.JwkService
}

var _ openapi.JwksControllerAPI = (*jwksController)(nil)

func NewJwksController(jwksService service.JwkService) openapi.JwksControllerAPI {
	return &jwksController{jwksService}
}

func (j *jwksController) GetJwks(ctx *gin.Context) {
	activeJwks, err := j.JwksService.GetActiveJwks(ctx)
	if err != nil {
		slog.Error("Failed to fetch active jwk keys", "error", err)
		RespondWithError(ctx, http.StatusInternalServerError, openapi.UNKNOWN, err.Error())
		return
	}

	result := make([]openapi.Jwk, len(activeJwks))
	for i, activeJwk := range activeJwks {

		n := base64.RawURLEncoding.EncodeToString(activeJwk.PublicKey.N.Bytes())

		eBytes := big.NewInt(int64(activeJwk.PublicKey.E)).Bytes()
		if len(eBytes) < 4 {
			padding := make([]byte, 4-len(eBytes))
			eBytes = append(padding, eBytes...)
		}
		e := base64.RawURLEncoding.EncodeToString(eBytes)

		result[i] = openapi.Jwk{
			Kty: activeJwk.Kty,
			Kid: activeJwk.Id,
			Use: activeJwk.Use,
			Alg: activeJwk.Alg,
			N:   n,
			E:   e,
		}
	}

	ctx.JSON(http.StatusOK, openapi.Jwks{
		Keys: result,
	})
}
