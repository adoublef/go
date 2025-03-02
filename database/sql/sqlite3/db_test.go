// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite3_test

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	. "go.adoublef.dev/database/sql/sqlite3"
	"go.adoublef.dev/testing/is"
)

//go:embed all:testdata/*.up.sql
var embedFS embed.FS
var sqlFS, _ = NewFS(embedFS, "testdata")

func Test_Up(t *testing.T) {
	t.Run("OK", func(t *testing.T) {

		_, err := sqlFS.Up(context.TODO(), testFilename(t, "test.db"))
		is.OK(t, err) // (sql3.FS).Up
	})
}

func Test_DB_Tx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is a long test")
	}

	t.Run("OK", testRoundTrip(func(db *DB) {
		tx, err := db.Tx(context.TODO())
		is.OK(t, err) // (sql3.DB).Tx

		t.Cleanup(func() { tx.Rollback() })

		for i := range 5_000_000 {
			rid := uuid.Must(uuid.NewV7())
			_, err = tx.Exec(context.TODO(), `insert into tests (id, counter) values (?, ?)`, rid, i)
			is.OK(t, err) // (sql3.Tx).Exec

			if i%500_000 == 0 {
				err = tx.Commit()
				is.OK(t, err) // (sql3.Tx).Commit
				tx, err = db.Tx(context.TODO())
				is.OK(t, err) // (sql3.DB).Tx
			}
		}

		is.OK(t, tx.Commit()) // (sql3.Tx).Commit

		// find
		var rid uuid.UUID
		err = db.QueryRow(context.TODO(), `select id from tests order by id desc limit 1`).Scan(&rid)
		is.OK(t, err) // (sql3.DB).QueryRow
	}))
}

func testRoundTrip(f func(*DB)) func(*testing.T) {
	return func(t *testing.T) {
		db, err := sqlFS.Up(context.TODO(), t.TempDir()+"/test.db")
		if err != nil {
			t.Fatalf("sql3.Up: %v", err)
		}
		t.Cleanup(func() { db.Close() })
		f(db)
	}
}

func testFilename(t testing.TB, filename string) string {
	t.Helper()
	if os.Getenv("DEBUG") != "1" {
		return filepath.Join(t.TempDir(), filename)
	}
	_ = os.Remove(filename)
	return filepath.Join(filename)
}
