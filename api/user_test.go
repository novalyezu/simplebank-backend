package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	mockdb "github.com/novalyezu/simplebank-backend/db/mock"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	assert.NoError(t, err)

	return db.User{
		Username:       util.RandomString(6),
		HashedPassword: hashedPassword,
		Email:          fmt.Sprintf("%s@email.com", util.RandomString(6)),
		FullName:       util.RandomString(6),
	}, password
}

func requiredUserMatchBody(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	assert.NoError(t, err)

	assert.Equal(t, user, gotUser)
}

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          createUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: createUserRequest{
				Username: user.Username,
				Password: password,
				Email:    user.Email,
				FullName: user.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
						Username:       user.Username,
						HashedPassword: user.HashedPassword,
						Email:          user.Email,
						FullName:       user.FullName,
					}, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				requiredUserMatchBody(t, recorder.Body, user)
			},
		},
		{
			name: "BadRequest",
			body: createUserRequest{
				Username: "",
				Password: "",
				Email:    "",
				FullName: "",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: createUserRequest{
				Username: user.Username,
				Password: password,
				Email:    user.Email,
				FullName: user.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Constraint: "users_pkey"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, recorder.Code)
				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				err = json.Unmarshal(data, &resp)
				assert.NoError(t, err)

				assert.Contains(t, resp["error"], "username already exists")
			},
		},
		{
			name: "DuplicateEmail",
			body: createUserRequest{
				Username: user.Username,
				Password: password,
				Email:    user.Email,
				FullName: user.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Constraint: "users_email_key"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, recorder.Code)
				data, err := io.ReadAll(recorder.Body)
				assert.NoError(t, err)

				resp := gin.H{}
				err = json.Unmarshal(data, &resp)
				assert.NoError(t, err)

				assert.Contains(t, resp["error"], "email already exists")
			},
		},
		{
			name: "InternalServerError",
			body: createUserRequest{
				Username: user.Username,
				Password: password,
				Email:    user.Email,
				FullName: user.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
						Username:       user.Username,
						HashedPassword: user.HashedPassword,
						Email:          user.Email,
						FullName:       user.FullName,
					}, password)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
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

			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := "/users"
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
