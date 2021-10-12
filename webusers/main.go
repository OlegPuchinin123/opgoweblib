package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/oleg/opgoweblib"
)

const (
	DBNAME  = "/var/www/users.db"
	REGFAIL = "Регистрация провалена."
)

type WebUsers struct {
	users *opgoweblib.Users
}

func (wu *WebUsers) AllUsers() ([]byte, error) {
	var (
		e      error
		r      *sql.Rows
		stmt   *sql.Stmt
		q      string
		user1  opgoweblib.User
		result string
	)

	wu.users.M.RLock()
	defer wu.users.M.RUnlock()

	q = "SELECT login,email FROM Users;"
	stmt, e = wu.users.Db.Prepare(q)
	if e != nil {
		fmt.Fprintf(os.Stderr, "Prepare loose.\n")
		return nil, e
	}
	r, e = stmt.Query()
	if e != nil {
		return nil, e
	}

	for r.Next() {
		if r.Scan(&user1.Login, &user1.Email) != nil {
			break
		}
		result += user1.Login + " " + user1.Email + "<br>"
	}
	r.Close()
	return []byte(result), nil
}

func (wu *WebUsers) Check_user_form(h map[string]string) (bool, string) {
	if h["Pass"] != "Pass" {
		return false, REGFAIL + " Неправильная форма."
	}
	if h["Login"] == "" {
		return false, REGFAIL + " Не указан логин."

	}
	if h["Email"] == "" {
		return false, REGFAIL + " Не указан E-mail."
	}
	if h["Password"] == "" {
		return false, REGFAIL + " Не указан пароль."
	}

	if h["Password"] != h["Password_again"] {
		return false, REGFAIL + " Пароли не совпадают."
	}

	return true, "Регистрация успешна."
}

func (wu *WebUsers) Add_user(w http.ResponseWriter, r *http.Request) {
	var (
		bresult bool
		e       error
		h       map[string]string
		result  string
		user1   *opgoweblib.User
	)

	h = make(map[string]string)
	h["Pass"] = r.PostFormValue("Pass")
	h["Login"] = r.PostFormValue("Login")
	h["Email"] = r.PostFormValue("Email")
	h["Password"] = r.PostFormValue("Password")
	h["Password_again"] = r.PostFormValue("Password_again")

	bresult, result = wu.Check_user_form(h)
	if !bresult {
		w.Write([]byte(result))
	} else {
		user1, _ = wu.users.Find_user(h["Login"])
		if user1 != nil {
			w.Write([]byte(REGFAIL + " Пользователь существует !"))
			return
		}
		e = wu.users.Add_user(h["Login"], h["Password"], h["Email"])
		if e == nil {
			w.Write([]byte("Регистрация успешна !"))
		} else {
			w.Write([]byte("Ошибка регистрации: " + e.Error()))
		}
	}
}

func (wu *WebUsers) handle_info_page(w http.ResponseWriter, r *http.Request) {
	var (
		buf []byte
		e   error
	)
	w.Write([]byte("Пользователи: <br>"))
	buf, e = wu.AllUsers()
	if e != nil {
		w.Write([]byte("AllUser eror " + e.Error() + "\n"))
	} else {
		w.Write(buf)
	}
}

func (wu *WebUsers) handle_register_page(w http.ResponseWriter, r *http.Request) {
	var (
		form string
	)
	form += "Регистрация:<br>"
	form += "<form action=\"/user_register\" method=\"POST\">\n"
	form += "Логин: <input type=\"text\" name=\"Login\" value=\"\" /><br>"
	form += "Email: <input type=\"text\" name=\"Email\" value=\"\" /><br>"
	form += "Пароль: <input type=\"password\" name=\"Password\" value=\"\" /><br>"
	form += "Повтор пароля: <input type=\"password\" name=\"Password_again\" value=\"\" /><br>"
	form += "<input type=\"hidden\" name=\"Pass\" value=\"Pass\" /><br>"
	form += "<input type=\"submit\" name=\"Submit\" value=\"Подтвердить\" /><br>"
	form += "</form>"
	w.Write([]byte(form))
}

func (wu *WebUsers) send_user_error(w http.ResponseWriter, r *http.Request, e string) {
	w.Write([]byte(opgoweblib.HTML_Start))
	w.Write([]byte("Вход не выполнен ! " + e + " <a href=\"/user_login\">назад</a>"))
	w.Write([]byte(opgoweblib.HTML_End))
}

func (wu *WebUsers) handle_logout_page(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Set-Cookie", "User=")
	w.Write([]byte(opgoweblib.HTML_Start))
	w.Write([]byte("<a href=\"/user_login\">Назад</a>"))
	w.Write([]byte(opgoweblib.HTML_End))
}

func (wu *WebUsers) handle_login_page(w http.ResponseWriter, r *http.Request) {
	var (
		e                     error
		form                  string
		login, password, pass string
		user1                 *opgoweblib.User
	)

	if r.Method == "POST" {
		login = r.PostFormValue("Login")
		password = r.PostFormValue("Password")
		pass = r.PostFormValue("Pass")
		if pass != "Pass" {
			wu.send_user_error(w, r, "Pass != Pass")
			return
		}
		user1, e = wu.users.Find_user(login)
		if (user1 == nil) || (e != nil) {
			wu.send_user_error(w, r, "Пара пользователь + пароль неверна.")
			return
		}
		if user1.Password != wu.users.CodeLoginPass(user1.Login, password) {
			wu.send_user_error(w, r, "Пара пользователь + пароль неверна.")
			return
		}
		w.Header().Add("Set-Cookie", "User="+login)
		w.Write([]byte(opgoweblib.HTML_Start))
		w.Write([]byte("Вы удачно зашли ! <a href=\"/user_login\">назад</a>"))
		w.Write([]byte(opgoweblib.HTML_End))
	} else if r.Method == "GET" {
		w.Write([]byte(opgoweblib.HTML_Start))
		login = opgoweblib.Check_cookie("User", r)
		if login != "" {
			w.Write([]byte("Ваш логин: " + login + "<br />" + "<a href=\"/user_logout\">Выйти</a>"))
		} else {
			//w.Write([]byte("Вы не вошли.<br>"))
			w.Write([]byte("Вход"))
			form += "<br><form action=\"/user_login\" method=\"POST\" />"
			form += "Логин: <input type=\"text\" name=\"Login\" value=\"\" />  <br />"
			form += "Пароль: <input type=\"password\" name=\"Password\" value=\"\" /> <br />"
			form += "<input type=\"submit\" name=\"Submit\" value=\"Отправить\" /> <br />"
			form += "<input type=\"hidden\" name=\"Pass\" value=\"Pass\" /> <br />"
			form += "</form>"
			w.Write([]byte(form))
		}
		w.Write([]byte(opgoweblib.HTML_End))
	}
}

func (wu *WebUsers) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/user_login" {
		wu.handle_login_page(w, r)
		return
	}

	if r.URL.Path == "/user_logout" {
		wu.handle_logout_page(w, r)
		return
	}

	w.Write([]byte(opgoweblib.HTML_Start))
	if r.URL.Path == "/user_register" {
		if r.Method == "GET" {
			wu.handle_register_page(w, r)
		} else if r.Method == "POST" {
			wu.Add_user(w, r)
		}
	}

	if r.URL.Path == "/users_info" {
		wu.handle_info_page(w, r)
	}
	w.Write([]byte(opgoweblib.HTML_End))
}

func main() {
	var (
		appSignal chan os.Signal
		b         bool
		e         error
		wu        *WebUsers
		port      *string
		//user1 *opgoweblib.User
	)

	port = flag.String("port", ":3812", "")
	flag.Parse()

	wu = new(WebUsers)
	wu.users = opgoweblib.NewUsers()
	wu.users.DB_load(DBNAME)
	if len(os.Args) >= 2 {
		if os.Args[1] == "create_db" {
			wu.users.DB_create(DBNAME)
			return
		}
		if os.Args[1] == "test" {
			b, e = wu.users.Login("Oleg", "123")
			if b && (e == nil) {
				fmt.Printf("User login ok !\n")
			}
			return
		}
	}
	appSignal = make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		select {
		case <-appSignal:
			wu.users.Db.Close()
			fmt.Printf("Stopped.")
			os.Exit(0)
		}
	}()

	http.HandleFunc("/", wu.handler)
	http.ListenAndServe(*port, nil)
}
