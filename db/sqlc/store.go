package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	DeleteTransferTx(ctx context.Context, arg DeleteTransferTxParams) error
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		errRb := tx.Rollback()
		if errRb != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, errRb)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// to prevent deadlock error because of 2 or more processes concurrently update same row on same table at same time
		// we need to order/sort the queries update by ID ASC
		// example case:
		// go1 => goroutine1, go2 => goroutine2
		// go1 transfer money from account1 to account2 with ID account1.ID=1 and account2.ID=2
		// go2 transfer money from account2 to account1 with same ID as above
		// 1. go1 and go2 running concurrently and lets say go1 run first
		// 2. go1 update account1 balance first and will locked account1 row
		// 3. go2 try to update account1 balance first also and it will be blocked by go1 and waiting until tx commit or rollback
		// 4. go1 continue the process update account2 balance and commit, the lock is released
		// 5. go2 can continue the process to update account1 balance and then account2 and then commit
		// what if we don't order/sort the queries update by ID ASC? deadlock will happen, but how?
		// see on steps 3, imagine go2 try to update account2 first instead of account1
		// the process of go2 will not be blocked, lets see:
		// 1. go1 and go2 running concurrently and lets say go1 run first
		// 2. go1 update account1 balance first and will locked account1 row
		// 3. go2 update account2 balance first and will locked account2 row
		// 4. go1 want to continue the process to update account2 balance, but account2 is locked and blocked by go2, go1 is waiting here
		// 5. go2 try to update account1 balance, but account1 is locked and blocked by go1, go2 is waiting here
		// 6. go1 and go2 are waiting each other, so deadlock will happen, it is just because we don't order/sort the queries.
		// Order Queries MATTERS!!!!
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = moveBalance(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = moveBalance(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func moveBalance(
	ctx context.Context,
	q *Queries,
	account1ID int64,
	amount1 int64,
	account2ID int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     account1ID,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     account2ID,
		Amount: amount2,
	})
	return
}

type DeleteTransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
}

// for testing purpose
func (store *SQLStore) DeleteTransferTx(ctx context.Context, arg DeleteTransferTxParams) error {
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		err = q.DeleteTransfer(ctx, DeleteTransferParams(arg))
		if err != nil {
			return err
		}

		err = q.DeleteEntryByAccountID(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		err = q.DeleteEntryByAccountID(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}
