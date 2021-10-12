package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oleg/opgoweblib"
)

func main() {
	var (
		sqlStmt string
	)
	opgoweblib.SayHello()

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt = `
	create table foo (id integer not null primary key, name text);
	delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	db.Close()
}
