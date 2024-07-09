package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	mockdb "github.com/novalyezu/simplebank-backend/db/mock"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/token"
	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateTransfer(t *testing.T) {
	user1, _ := randomUser(t)
	account1 := randomAccount(user1.Username)
	account1.Balance = 1000
	account1.Currency = util.IDR

	user2, _ := randomUser(t)
	account2 := randomAccount(user2.Username)
	account2.Balance = 500
	account2.Currency = util.IDR

	amount := int64(100)

	transferTxResult := db.TransferTxResult{
		Transfer: db.Transfer{
			ID:            util.RandomInt(1, 99),
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        amount,
		},
		FromAccount: account1,
		ToAccount:   account2,
		FromEntry: db.Entry{
			ID:        util.RandomInt(1, 99),
			AccountID: account1.ID,
			Amount:    -amount,
		},
		ToEntry: db.Entry{
			ID:        util.RandomInt(1, 99),
			AccountID: account2.ID,
			Amount:    amount,
		},
	}

	testCases := []struct {
		name          string
		body          transferRequest
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: transferRequest{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
				Currency:      util.IDR,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, err := tokenMaker.CreateToken(user1.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: account1.ID,
						ToAccountID:   account2.ID,
						Amount:        amount,
					})).
					Times(1).
					Return(transferTxResult, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				var gotTransferTxResult db.TransferTxResult
				err = json.Unmarshal(data, &gotTransferTxResult)
				assert.NoError(t, err)

				assert.Equal(t, transferTxResult, gotTransferTxResult)
			},
		},
		{
			name: "BadRequest",
			body: transferRequest{
				FromAccountID: 0,
				ToAccountID:   0,
				Amount:        0,
				Currency:      "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, err := tokenMaker.CreateToken(user1.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TransferToSameAccount",
			body: transferRequest{
				FromAccountID: account1.ID,
				ToAccountID:   account1.ID,
				Amount:        amount,
				Currency:      util.IDR,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, err := tokenMaker.CreateToken(user1.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				err = json.Unmarshal(data, &resp)
				assert.NoError(t, err)

				assert.Contains(t, resp["error"], "cannot transfer to same account")
			},
		},
		{
			name: "TransferFromOtherAccount",
			body: transferRequest{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
				Currency:      util.IDR,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, err := tokenMaker.CreateToken(user2.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				err = json.Unmarshal(data, &resp)
				assert.NoError(t, err)

				assert.Contains(t, resp["error"], "cannot transfer from other account")
			},
		},
		{
			name: "InvalidCurrency",
			body: transferRequest{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
				Currency:      util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, err := tokenMaker.CreateToken(user1.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				err = json.Unmarshal(data, &resp)
				assert.NoError(t, err)

				assert.Contains(t, resp["error"], "currency not valid")
			},
		},
		{
			name: "InvalidBalance",
			body: transferRequest{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        int64(1500),
				Currency:      util.IDR,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, err := tokenMaker.CreateToken(user1.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				store.
					EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				err = json.Unmarshal(data, &resp)
				assert.NoError(t, err)

				assert.Contains(t, resp["error"], "balance not valid")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := newServerTest(t, store)
			recorder := httptest.NewRecorder()
			url := "/transfers"
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
