package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oleg/opgoweblib"
)

const DBNAME = "/var/www/links.db"

type Server struct {
	my_mutex sync.RWMutex
	db       *sql.DB
}

func (serv *Server) AddLink(name, link string, info string, w http.ResponseWriter) error {
	var (
		e    error
		stmt *sql.Stmt
		q    string
	)

	serv.my_mutex.Lock()
	defer serv.my_mutex.Unlock()

	q = "INSERT INTO links (name,link)\n"
	q += "VALUES (?, ?);"
	stmt, e = serv.db.Prepare(q)
	if e != nil {
		return e
	}
	_, e = stmt.Exec(name, link)
	return e
}

func (serv *Server) db_links() []byte {
	var (
		err        error
		sqlStmt    string
		s          string
		rows       *sql.Rows
		name, link string
		id         int
	)

	serv.my_mutex.RLock()
	defer serv.my_mutex.RUnlock()

	sqlStmt = `
	SELECT * FROM links;	
	`
	rows, err = serv.db.Query(sqlStmt)
	defer rows.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q: %s\n", err, sqlStmt)
		return nil
	}
	s = ""
	for rows.Next() {
		err = rows.Scan(&id, &name, &link)
		if err != nil {
			return nil
		}
		link = fmt.Sprintf("<a href=\"%s\">%s</a>\n<br>", link, name)
		s += link
	}
	return []byte(s)
}

func (serv *Server) db_load(fname string) error {
	var (
		e error
	)
	serv.db, e = sql.Open("sqlite3", fname)
	return e
}

func (serv *Server) create_db(fname string) error {
	var (
		db      *sql.DB
		err     error
		sqlStmt string
	)
	_, err = os.Stat(fname)
	if err == nil {
		os.Remove(fname)
	}
	db, err = sql.Open("sqlite3", DBNAME)
	if err != nil {
		return nil
	}
	defer db.Close()

	sqlStmt = `
	CREATE TABLE links (id integer not null primary key, name text, link text);
	DELETE from links;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func (serv *Server) handler(w http.ResponseWriter, r *http.Request) {
	var (
		form       string
		links_html []byte
	)

	w.Write([]byte(opgoweblib.HTML_Start))
	if r.URL.Path == "/links_form" {
		if r.Method == "POST" {
			r.ParseForm()
			if r.PostFormValue("Pass") == "Pass" {
				serv.AddLink(r.PostFormValue("Name"), r.PostFormValue("Link"),
					r.PostFormValue("Info"), w)
			}
		}
		w.Write([]byte("Ссылки<br>\n"))
		links_html = serv.db_links()
		w.Write(links_html)
		form = "<form action=\"/links_form\" method=\"POST\"><br>\n"
		form += "<input type=\"text\" name=\"Name\" value=\"Имя\" /> <br>"
		form += "<input type=\"text\" name=\"Link\" value=\"Ссылка\" /> <br>"
		form += "<input type=\"submit\" name=\"s\" value=\"Добавить.\"/>\n"
		form += "<input type=\"hidden\" name=\"Pass\" value=\"Pass\">"
		form += "</form>\n"
		w.Write([]byte(form))
	}

	if r.URL.Path == "/links" {
		w.Write([]byte("Ссылки<br>\n"))
		links_html = serv.db_links()
		w.Write(links_html)
	}
	w.Write([]byte(opgoweblib.HTML_End))
}

func main() {
	var (
		appSignal   chan os.Signal
		serv        *Server
		addr_string *string
	)

	serv = new(Server)
	serv.db_load(DBNAME)

	addr_string = flag.String("port", ":8080", "")
	flag.Parse()

	if len(os.Args) == 2 {
		if os.Args[1] == "create_db" {
			serv.create_db("/var/www/links.db")
			return
		} else if os.Args[1] == "show_all" {
			os.Stdout.Write(serv.db_links())
			return
		}
	}
	appSignal = make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		select {
		case <-appSignal:
			serv.db.Close()
			fmt.Printf("Stopped.")
			os.Exit(0)
		}
	}()

	http.HandleFunc("/", serv.handler)
	http.ListenAndServe(*addr_string, nil)
}
