package opgoweblib

import (
	"database/sql"
	"sync"
)

type GDB struct {
	Db *sql.DB
	M  sync.RWMutex
}

func (u *GDB) DB_load(fname string) error {
	var (
		e error
	)
	u.Db, e = sql.Open("sqlite3", fname)
	if e != nil {
		return e
	}
	return nil
}

func (u *GDB) DB_close() {
	u.Db.Close()
}
