package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/token"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(c *gin.Context) {
	var body transferRequest

	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if body.FromAccountID == body.ToAccountID {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("cannot transfer to same account")))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)
	fromAccount := server.getAccountByID(c, body.FromAccountID)
	toAccount := server.getAccountByID(c, body.ToAccountID)

	if authPayload.Username != fromAccount.Owner {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("cannot transfer from other account")))
		return
	}

	if !server.isValidCurrency(c, fromAccount, body.Currency) {
		return
	}
	if !server.isValidCurrency(c, toAccount, body.Currency) {
		return
	}
	if !server.isValidBalance(c, fromAccount, body.Amount) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: body.FromAccountID,
		ToAccountID:   body.ToAccountID,
		Amount:        body.Amount,
	}

	result, err := server.store.TransferTx(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) getAccountByID(c *gin.Context, accountID int64) db.Account {
	account, err := server.store.GetAccount(c, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return db.Account{}
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return db.Account{}
	}
	return account
}

func (server *Server) isValidCurrency(c *gin.Context, account db.Account, currency string) bool {
	if account.Currency != currency {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("currency not valid [%s], %s vs %s", account.Owner, account.Currency, currency)))
		return false
	}
	return true
}

func (server *Server) isValidBalance(c *gin.Context, account db.Account, amount int64) bool {
	if account.Balance < amount {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("balance not valid [%s], %d vs %d", account.Owner, account.Balance, amount)))
		return false
	}
	return true
}
