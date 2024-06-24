package db

import (
	"context"
	"testing"

	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

const accEntryPrefix = "entry_test_"

func createRandomEntry(t *testing.T, account Account) Entry {
	ctx := context.Background()

	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomInt(0, 20),
	}

	entry, err := testQueries.CreateEntry(ctx, arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry)

	assert.Equal(t, arg.AccountID, entry.AccountID)
	assert.Equal(t, arg.Amount, entry.Amount)

	assert.NotZero(t, entry.ID)
	assert.NotZero(t, entry.CreatedAt)

	return entry
}

func deleteTestingEntry(ctx context.Context, accountID int64) {
	testQueries.deleteEntryByAccountID(ctx, accountID)
}

func TestCreateEntry(t *testing.T) {
	ctx := context.Background()
	account := createRandomAccount(t, accEntryPrefix)
	createRandomEntry(t, account)
	defer deleteTestingAccount(ctx, accEntryPrefix)
	defer deleteTestingEntry(ctx, account.ID)
}

func TestGetEntry(t *testing.T) {
	ctx := context.Background()
	account := createRandomAccount(t, accEntryPrefix)
	newEntry := createRandomEntry(t, account)
	defer deleteTestingAccount(ctx, accEntryPrefix)
	defer deleteTestingEntry(ctx, account.ID)

	entry, err := testQueries.GetEntry(ctx, newEntry.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry)

	assert.Equal(t, newEntry.ID, entry.ID)
	assert.Equal(t, newEntry.AccountID, entry.AccountID)
	assert.Equal(t, newEntry.Amount, entry.Amount)
}

func TestListEntries(t *testing.T) {
	ctx := context.Background()
	account := createRandomAccount(t, accEntryPrefix)

	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	defer deleteTestingAccount(ctx, accEntryPrefix)
	defer deleteTestingEntry(ctx, account.ID)

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}
	entries, err := testQueries.ListEntries(ctx, arg)
	assert.NoError(t, err)
	assert.Len(t, entries, 5)

	for _, entry := range entries {
		assert.NotEmpty(t, entry)
	}
}
