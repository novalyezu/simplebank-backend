package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/novalyezu/simplebank-backend/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationType       = "bearer"
	authorizationPayloadKey = "auth_payload"
)

var (
	ErrTokenIsNotProvided  = errors.New("token is not provided")
	ErrTokenIsInvalid      = errors.New("token is invalid")
	ErrUnsupportedAuthType = errors.New("unsupported authorization type")
)

func authMiddleware(token token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrTokenIsNotProvided))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) != 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrTokenIsInvalid))
			return
		}

		if strings.ToLower(fields[0]) != authorizationType {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrUnsupportedAuthType))
			return
		}

		accessToken := fields[1]
		payload, err := token.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
