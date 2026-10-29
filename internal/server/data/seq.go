package data

import (
	"context"
	"database/sql"
	"errors"
)

// SeqLock hands out a user's sync seq values inside a transaction. Every write
// to a user's todos takes the next value and stores it in the row's seq
// column, so clients can ask for "every row with seq > N" to find what changed.
//
// LockSeq locks the user's row with SELECT ... FOR UPDATE. Postgres holds that
// lock until the transaction commits or rolls back, and any other transaction
// that tries to lock the same row waits. So writes for one user run one at a
// time, and their seq values commit in increasing order. Writes for different
// users lock different rows and don't block each other.
//
// A global Postgres sequence would not be enough. Two transactions could take
// seq 7 and 8 and commit in the opposite order. A client that syncs between
// those commits would see 8, record last_seq = 8, and never receive 7.
//
// Typical use:
//
//	tx, err := db.BeginTx(ctx, nil)
//	// handle err, defer tx.Rollback()
//	seq, err := LockSeq(ctx, tx, userID)
//	// handle err
//	for each write {
//		// ... UPDATE todos SET seq = seq.Next(), ...
//	}
//	err = seq.Save(ctx)
//	// handle err
//	err = tx.Commit()
//
// A SeqLock is only valid for the transaction it was created in, and is not
// safe for concurrent use.
type SeqLock struct {
	tx     *sql.Tx
	userID int64
	start  int64 // users.sync_seq when the lock was taken
	last   int64 // the most recent value returned by Next
}

// LockSeq locks the user's row in the users table for the rest of the
// transaction and reads the current sync_seq. It returns ErrRecordNotFound if
// there is no user with the given ID.
//
// Callers should take this lock before touching any of the user's todo rows,
// so that every writer acquires locks in the same order.
func LockSeq(ctx context.Context, tx *sql.Tx, userID int64) (*SeqLock, error) {
	query := `SELECT sync_seq FROM users WHERE id = $1 FOR UPDATE`

	var current int64
	err := tx.QueryRowContext(ctx, query, userID).Scan(&current)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &SeqLock{tx: tx, userID: userID, start: current, last: current}, nil
}

// Next returns the next seq value for the user. Each call returns a value one
// greater than the previous call. Nothing is written to the database until
// Save is called.
func (s *SeqLock) Next() int64 {
	s.last++
	return s.last
}

// Last returns the most recent value returned by Next. If Next hasn't been
// called, it returns the user's sync_seq as of when the lock was taken.
func (s *SeqLock) Last() int64 {
	return s.last
}

// Save writes the most recent seq value back to users.sync_seq. It must be
// called before the transaction commits. If Next was never called, Save does
// nothing.
func (s *SeqLock) Save(ctx context.Context) error {
	if s.last == s.start {
		return nil
	}

	query := `UPDATE users SET sync_seq = $1 WHERE id = $2`
	_, err := s.tx.ExecContext(ctx, query, s.last, s.userID)
	if err != nil {
		return err
	}

	s.start = s.last
	return nil
}

// WithSeqLock runs fn in a new transaction that holds the user's seq lock.
// If fn returns nil, the final seq value is saved and the transaction is
// committed. If fn returns an error, the transaction is rolled back and the
// error is returned as-is.
func WithSeqLock(ctx context.Context, db *sql.DB, userID int64, fn func(tx *sql.Tx, seq *SeqLock) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Rollback is a no-op after a successful commit.
	defer tx.Rollback()

	seq, err := LockSeq(ctx, tx, userID)
	if err != nil {
		return err
	}

	err = fn(tx, seq)
	if err != nil {
		return err
	}

	err = seq.Save(ctx)
	if err != nil {
		return err
	}

	return tx.Commit()
}
