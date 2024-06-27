package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const storeTestPrefix = "store_test_"

func TestTransferTx(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testDB)
	account1 := createRandomAccount(t, storeTestPrefix)
	account2 := createRandomAccount(t, storeTestPrefix)

	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	for i := 0; i < n; i++ {
		ctx := context.Background()
		err := <-errs
		assert.NoError(t, err)

		result := <-results
		assert.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		assert.NotEmpty(t, transfer)
		assert.Equal(t, account1.ID, transfer.FromAccountID)
		assert.Equal(t, account2.ID, transfer.ToAccountID)
		assert.Equal(t, amount, transfer.Amount)
		assert.NotZero(t, transfer.ID)
		assert.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(ctx, transfer.ID)
		assert.NoError(t, err)

		// check from account entry
		fromEntry := result.FromEntry
		assert.NotEmpty(t, fromEntry)
		assert.Equal(t, account1.ID, fromEntry.AccountID)
		assert.Equal(t, -amount, fromEntry.Amount)
		assert.NotZero(t, fromEntry.ID)
		assert.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(ctx, fromEntry.ID)
		assert.NoError(t, err)

		// check to account entry
		toEntry := result.ToEntry
		assert.NotEmpty(t, toEntry)
		assert.Equal(t, account2.ID, toEntry.AccountID)
		assert.Equal(t, amount, toEntry.Amount)
		assert.NotZero(t, toEntry.ID)
		assert.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(ctx, toEntry.ID)
		assert.NoError(t, err)

		// check from account balance
		fromAccount := result.FromAccount
		expectedFromAccountBalance := account1.Balance - (int64(i+1) * amount)
		assert.NotEmpty(t, fromAccount)
		assert.Equal(t, account1.ID, fromAccount.ID)
		assert.Equal(t, expectedFromAccountBalance, fromAccount.Balance)

		// check to account balance
		toAccount := result.ToAccount
		expectedToAccountBalance := account2.Balance + (int64(i+1) * amount)
		assert.NotEmpty(t, toAccount)
		assert.Equal(t, account2.ID, toAccount.ID)
		assert.Equal(t, expectedToAccountBalance, toAccount.Balance)
	}

	// check for final update
	updatedAccount1, err := store.GetAccount(ctx, account1.ID)
	expectedAccount1Balance := account1.Balance - (int64(n) * amount)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccount1Balance, updatedAccount1.Balance)

	updatedAccount2, err := store.GetAccount(ctx, account2.ID)
	expectedAccount2Balance := account2.Balance + (int64(n) * amount)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccount2Balance, updatedAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testDB)
	account1 := createRandomAccount(t, storeTestPrefix)
	account2 := createRandomAccount(t, storeTestPrefix)

	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		// if odd number, reverse the transfers
		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			ctx := context.Background()
			_, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		assert.NoError(t, err)
	}

	// check for final update
	updatedAccount1, err := store.GetAccount(ctx, account1.ID)
	assert.NoError(t, err)
	assert.Equal(t, account1.Balance, updatedAccount1.Balance)

	updatedAccount2, err := store.GetAccount(ctx, account2.ID)
	assert.NoError(t, err)
	assert.Equal(t, account2.Balance, updatedAccount2.Balance)
}
