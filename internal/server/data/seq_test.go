package data

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/kvnloughead/godo/internal/assert"
)

// TestSeqLock checks that LockSeq hands out consecutive values and that only
// a saved, committed transaction changes users.sync_seq.
func TestSeqLock(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	ctx := context.Background()

	t.Run("Next and Save", func(t *testing.T) {
		start := syncSeqFor(t, db, user.ID)

		err := WithSeqLock(ctx, db, user.ID, func(tx *sql.Tx, seq *SeqLock) error {
			assert.Equal(t, seq.Last(), start)
			assert.Equal(t, seq.Next(), start+1)
			assert.Equal(t, seq.Next(), start+2)
			assert.Equal(t, seq.Next(), start+3)
			assert.Equal(t, seq.Last(), start+3)
			return nil
		})
		assert.IsNil(t, err)
		assert.Equal(t, syncSeqFor(t, db, user.ID), start+3)
	})

	t.Run("Rollback discards values", func(t *testing.T) {
		start := syncSeqFor(t, db, user.ID)
		errBoom := errors.New("boom")

		err := WithSeqLock(ctx, db, user.ID, func(tx *sql.Tx, seq *SeqLock) error {
			seq.Next()
			seq.Next()
			return errBoom
		})
		if !errors.Is(err, errBoom) {
			t.Fatalf("got %v; want %v", err, errBoom)
		}
		assert.Equal(t, syncSeqFor(t, db, user.ID), start)
	})

	t.Run("Save without Next is a no-op", func(t *testing.T) {
		start := syncSeqFor(t, db, user.ID)

		err := WithSeqLock(ctx, db, user.ID, func(tx *sql.Tx, seq *SeqLock) error {
			return nil
		})
		assert.IsNil(t, err)
		assert.Equal(t, syncSeqFor(t, db, user.ID), start)
	})

	t.Run("Unknown user", func(t *testing.T) {
		err := WithSeqLock(ctx, db, -1, func(tx *sql.Tx, seq *SeqLock) error {
			t.Error("fn should not be called for an unknown user")
			return nil
		})
		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("got %v; want %v", err, ErrRecordNotFound)
		}
	})
}

// TestSeqLockBlocks checks that a second transaction for the same user waits
// for the first to commit, and then sees the first transaction's saved value.
func TestSeqLockBlocks(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	ctx := context.Background()

	start := syncSeqFor(t, db, user.ID)

	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx1.Rollback()

	seq1, err := LockSeq(ctx, tx1, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	seq1.Next()

	// The second transaction should block in LockSeq until tx1 commits.
	got := make(chan int64, 1)
	errs := make(chan error, 1)
	go func() {
		err := WithSeqLock(ctx, db, user.ID, func(tx *sql.Tx, seq *SeqLock) error {
			got <- seq.Next()
			return nil
		})
		errs <- err
	}()

	select {
	case v := <-got:
		t.Fatalf("second transaction got seq %d while the first held the lock", v)
	case <-time.After(300 * time.Millisecond):
	}

	err = seq1.Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = tx1.Commit()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case v := <-got:
		assert.Equal(t, v, start+2)
	case <-time.After(5 * time.Second):
		t.Fatal("second transaction never acquired the lock")
	}
	assert.IsNil(t, <-errs)
	assert.Equal(t, syncSeqFor(t, db, user.ID), start+2)
}
