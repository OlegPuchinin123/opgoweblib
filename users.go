package opgoweblib

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
)

type User struct {
	Login    string
	Password string
	Email    string
	Id       int
}

type Users struct {
	GDB
}

func NewUsers() *Users {
	var (
		u *Users
	)
	u = new(Users)
	return u
}

func (u *Users) DB_create(fname string) error {
	var (
		db      *sql.DB
		e       error
		sqlStmt string
	)
	os.Remove(fname)
	db, e = sql.Open("sqlite3", fname)
	if e != nil {
		return e
	}
	defer db.Close()
	sqlStmt = `
	CREATE TABLE Users (id integer not null primary key, login text, password text, email text);
	DELETE from Users;
	`
	_, e = db.Exec(sqlStmt)
	if e != nil {
		return e
	}

	return nil
}

func (u *Users) Add_user(login, password, email string) error {
	var (
		e    error
		q    string
		stmt *sql.Stmt
	)

	u.M.Lock()
	defer u.M.Unlock()

	q = "INSERT INTO Users (login, password, email)\n"
	q += "VALUES (?, ?, ?);"
	stmt, e = u.Db.Prepare(q)
	if e != nil {
		return e
	}
	_, e = stmt.Exec(login, u.CodeLoginPass(login, password), email)
	return e
}

func (u *Users) Update_user(login, newpassword, newemail string) error {
	var (
		e    error
		q    string
		stmt *sql.Stmt
	)

	u.M.Lock()
	defer u.M.Unlock()

	q = "UPDATE Users\nSET password=?, email=?\nWHERE login=?;"
	stmt, e = u.Db.Prepare(q)
	if e != nil {
		return e
	}
	_, e = stmt.Exec(u.CodeLoginPass(login, newpassword), newemail, login)
	return e
}

func (u *Users) Delete_user(id int) error {
	var (
		e    error
		q    string
		stmt *sql.Stmt
	)

	u.M.Lock()
	defer u.M.Unlock()

	q = "DELETE FROM Users\nWHERE id=?;"
	stmt, e = u.Db.Prepare(q)
	if e != nil {
		return e
	}
	_, e = stmt.Exec(id)
	return e
}

func (u *Users) Find_user(login string) (*User, error) {
	var (
		e     error
		q     string
		stmt  *sql.Stmt
		rows  *sql.Rows
		user1 *User
	)

	u.M.RLock()
	defer u.M.RUnlock()

	q = "SELECT * FROM Users WHERE login=?;"
	stmt, e = u.Db.Prepare(q)
	if e != nil {
		return nil, e
	}
	user1 = new(User)
	rows, e = stmt.Query(login)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	rows.Next()
	e = rows.Scan(&user1.Id, &user1.Login, &user1.Password, &user1.Email)
	if e != nil {
		return nil, e
	}
	return user1, nil
}

func (u *Users) CodeLoginPass(login, pass string) string {
	var (
		m [md5.Size]byte
	)
	m = md5.Sum([]byte(login + pass))
	return fmt.Sprintf("%x", m)
}

func (u *Users) Login(login, pass string) (bool, error) {
	var (
		e     error
		user1 *User
	)
	user1, e = u.Find_user(login)
	if e != nil {
		return false, e
	}
	if user1.Password == u.CodeLoginPass(login, pass) {
		return true, nil
	}
	return false, nil
}
