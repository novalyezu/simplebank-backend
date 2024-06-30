package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

const accPrefix = "acc_test_"

func createRandomAccount(t *testing.T, prefix string) Account {
	ctx := context.Background()
	arg := CreateAccountParams{
		Owner:    prefix + util.RandomString(6),
		Balance:  util.RandomInt(0, 1000),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, account)

	assert.Equal(t, arg.Owner, account.Owner)
	assert.Equal(t, arg.Balance, account.Balance)
	assert.Equal(t, arg.Currency, account.Currency)

	assert.NotZero(t, account.ID)
	assert.NotZero(t, account.CreatedAt)

	return account
}

func deleteTestingAccount(ctx context.Context, prefix string) {
	testQueries.DeleteAccountByOwnerLike(ctx, prefix)
}

func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	createRandomAccount(t, accPrefix)
	defer deleteTestingAccount(ctx, accPrefix)
}

func TestGetAccount(t *testing.T) {
	ctx := context.Background()
	newAccount := createRandomAccount(t, accPrefix)
	defer deleteTestingAccount(ctx, accPrefix)

	account, err := testQueries.GetAccount(ctx, newAccount.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, account)

	assert.Equal(t, newAccount.ID, account.ID)
	assert.Equal(t, newAccount.Owner, account.Owner)
	assert.Equal(t, newAccount.Balance, account.Balance)
	assert.Equal(t, newAccount.Currency, account.Currency)
}

func TestUpdateAccount(t *testing.T) {
	ctx := context.Background()
	newAccount := createRandomAccount(t, accPrefix)
	defer deleteTestingAccount(ctx, accPrefix)

	arg := UpdateAccountParams{
		ID:      newAccount.ID,
		Balance: util.RandomInt(0, 1000),
	}

	updatedAccount, err := testQueries.UpdateAccount(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedAccount)

	assert.Equal(t, newAccount.ID, updatedAccount.ID)
	assert.Equal(t, newAccount.Owner, updatedAccount.Owner)
	assert.Equal(t, arg.Balance, updatedAccount.Balance)
	assert.Equal(t, newAccount.Currency, updatedAccount.Currency)
}

func TestDeleteAccount(t *testing.T) {
	ctx := context.Background()
	newAccount := createRandomAccount(t, accPrefix)
	defer deleteTestingAccount(ctx, accPrefix)

	err := testQueries.DeleteAccount(ctx, newAccount.ID)
	assert.NoError(t, err)

	checkAccount, err := testQueries.GetAccount(ctx, newAccount.ID)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
	assert.Empty(t, checkAccount)
}

func TestListAccounts(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		createRandomAccount(t, accPrefix)
	}
	defer deleteTestingAccount(ctx, accPrefix)

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}
	accounts, err := testQueries.ListAccounts(ctx, arg)
	assert.NoError(t, err)
	assert.Len(t, accounts, 5)

	for _, account := range accounts {
		assert.NotEmpty(t, account)
	}
}
