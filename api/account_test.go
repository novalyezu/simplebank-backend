package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/novalyezu/simplebank-backend/db/mock"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/token"
	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(0, 100),
		Owner:    owner,
		Balance:  util.RandomInt(0, 100),
		Currency: util.RandomCurrency(),
	}
}

func requiredAccountMatchBody(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	assert.NoError(t, err)

	assert.Equal(t, account, gotAccount)
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				requiredAccountMatchBody(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			// build stubs
			tc.buildStubs(store)

			// start test
			server := newServerTest(t, store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountID)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccount(t *testing.T) {
	var accounts []db.Account
	user, _ := randomUser(t)

	n := 5
	for i := 0; i < n; i++ {
		accounts = append(accounts, randomAccount(user.Username))
	}

	testCases := []struct {
		name          string
		queryParams   listAccountRequest
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			queryParams: listAccountRequest{Page: 1, Limit: 5},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  5,
						Offset: 0,
					}).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				var resAccounts []db.Account
				err = json.Unmarshal(data, &resAccounts)
				assert.NoError(t, err)

				assert.Equal(t, http.StatusOK, recorder.Code)
				assert.Len(t, resAccounts, n)
			},
		},
		{
			name:        "BadRequest",
			queryParams: listAccountRequest{Page: 0, Limit: 0},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0).
					Return([]db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:        "InternalServerError",
			queryParams: listAccountRequest{Page: 1, Limit: 5},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  5,
						Offset: 0,
					}).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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
			url := fmt.Sprintf("/accounts?page=%d&limit=%d", tc.queryParams.Page, tc.queryParams.Limit)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		body          createAccountRequest
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: createAccountRequest{Currency: account.Currency},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateAccount(gomock.Any(), db.CreateAccountParams{
						Owner:    account.Owner,
						Currency: account.Currency,
						Balance:  0,
					}).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				requiredAccountMatchBody(t, recorder.Body, account)
			},
		},
		{
			name: "BadRequest",
			body: createAccountRequest{Currency: ""},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: createAccountRequest{Currency: account.Currency},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(user.Username, time.Minute)
				assert.NoError(t, err)
				request.Header.Add(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, token))
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateAccount(gomock.Any(), db.CreateAccountParams{
						Owner:    account.Owner,
						Currency: account.Currency,
						Balance:  0,
					}).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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
			url := "/accounts"
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
