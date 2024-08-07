package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/novalyezu/simplebank-backend/token"
	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				username := util.RandomString(6)
				token, err := tokenMaker.CreateToken(username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "TokenIsNotProvided",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				json.Unmarshal(data, &resp)

				assert.Contains(t, resp["error"], "token is not provided")
			},
		},
		{
			name: "TokenIsInvalid",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", "", ""))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				json.Unmarshal(data, &resp)

				assert.Contains(t, resp["error"], "token is invalid")
			},
		},
		{
			name: "UnsupportedAuthType",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", "unsupported", "token"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				json.Unmarshal(data, &resp)

				assert.Contains(t, resp["error"], "unsupported authorization type")
			},
		},
		{
			name: "TokenIsExpired",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				username := util.RandomString(6)
				token, err := tokenMaker.CreateToken(username, -time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				json.Unmarshal(data, &resp)

				assert.Contains(t, resp["error"], "token is expired")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newServerTest(t, nil)

			authPath := "/"
			server.router.GET(authPath, authMiddleware(server.tokenMaker), func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{}) })

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
