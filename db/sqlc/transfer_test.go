package db

import (
	"context"
	"testing"

	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

const accTransferPrefix = "tf_test_"

func createRandomTransfer(t *testing.T, fromAccount Account, toAccount Account) Transfer {
	ctx := context.Background()
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomInt(0, 20),
	}

	transfer, err := testQueries.CreateTransfer(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, transfer)

	assert.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	assert.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	assert.Equal(t, arg.Amount, transfer.Amount)

	assert.NotZero(t, transfer.ID)
	assert.NotZero(t, transfer.CreatedAt)
	return transfer
}

func deleteTestingTransfer(ctx context.Context, fromAccountID int64, toAccountID int64) {
	testQueries.deleteTransfer(ctx, deleteTransferParams{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
	})
}

func TestCreateTransfer(t *testing.T) {
	ctx := context.Background()
	account1 := createRandomAccount(t, accTransferPrefix)
	account2 := createRandomAccount(t, accTransferPrefix)
	createRandomTransfer(t, account1, account2)

	defer deleteTestingAccount(ctx, accTransferPrefix)
	defer deleteTestingTransfer(ctx, account1.ID, account2.ID)
}

func TestGetTransfer(t *testing.T) {
	ctx := context.Background()
	account1 := createRandomAccount(t, accTransferPrefix)
	account2 := createRandomAccount(t, accTransferPrefix)
	newTransfer := createRandomTransfer(t, account1, account2)

	defer deleteTestingAccount(ctx, accTransferPrefix)
	defer deleteTestingTransfer(ctx, account1.ID, account2.ID)

	transfer, err := testQueries.GetTransfer(ctx, newTransfer.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, transfer)

	assert.Equal(t, newTransfer.ID, transfer.ID)
	assert.Equal(t, newTransfer.FromAccountID, transfer.FromAccountID)
	assert.Equal(t, newTransfer.ToAccountID, transfer.ToAccountID)
	assert.Equal(t, newTransfer.Amount, transfer.Amount)
}

func TestListTransfers(t *testing.T) {
	ctx := context.Background()
	account1 := createRandomAccount(t, accTransferPrefix)
	account2 := createRandomAccount(t, accTransferPrefix)

	for i := 0; i < 10; i++ {
		createRandomTransfer(t, account1, account2)
	}

	defer deleteTestingAccount(ctx, accTransferPrefix)
	defer deleteTestingTransfer(ctx, account1.ID, account2.ID)

	arg := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testQueries.ListTransfers(ctx, arg)
	assert.NoError(t, err)
	assert.Len(t, transfers, 5)

	for _, transfer := range transfers {
		assert.NotEmpty(t, transfer)
	}
}
