package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/token"
	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

func newServerTest(t *testing.T, store db.Store) *Server {
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	assert.NoError(t, err)

	server := NewServer(store, tokenMaker)
	return server
}

func TestMain(t *testing.M) {
	gin.SetMode(gin.TestMode)
	code := t.Run()
	os.Exit(code)
}
