package db

import (
	"context"
	"testing"

	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

const userPrefix = "user_test_"

func createRandomUser(t *testing.T, prefix string) User {
	ctx := context.Background()
	arg := CreateUserParams{
		Username:       prefix + util.RandomString(6),
		HashedPassword: util.RandomString(12),
		FullName:       util.RandomString(6),
		Email:          util.RandomEmail(6),
	}

	user, err := testQueries.CreateUser(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, user)

	assert.Equal(t, arg.Username, user.Username)
	assert.Equal(t, arg.HashedPassword, user.HashedPassword)
	assert.Equal(t, arg.FullName, user.FullName)
	assert.Equal(t, arg.Email, user.Email)

	assert.True(t, user.PasswordChangedAt.IsZero())
	assert.NotZero(t, user.CreatedAt)

	return user
}

func deleteTestingUser(ctx context.Context, prefix string) {
	testQueries.DeleteUserByUsernameLike(ctx, prefix)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	createRandomUser(t, userPrefix)
	defer deleteTestingUser(ctx, userPrefix)
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	newUser := createRandomUser(t, userPrefix)
	defer deleteTestingUser(ctx, userPrefix)

	user, err := testQueries.GetUser(ctx, newUser.Username)
	assert.NoError(t, err)
	assert.NotEmpty(t, user)

	assert.Equal(t, newUser.Username, user.Username)
	assert.Equal(t, newUser.HashedPassword, user.HashedPassword)
	assert.Equal(t, newUser.FullName, user.FullName)
	assert.Equal(t, newUser.Email, user.Email)
}
