package main

import (
	"flag"
	"net/http"

	"github.com/oleg/opgoweblib"
)

func main() {
	var (
		port *string
	)
	port = flag.String("port", ":3819", "")
	flag.Parse()

	http.HandleFunc("/", handler)
	http.ListenAndServe(*port, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/user_info" {
		//w.Header().Add("Set-Cookie", "status=looser")
		//w.Header().Add("Set-Cookie", "status=")

		w.Write([]byte(opgoweblib.HTML_Start))
		w.Write([]byte(r.RemoteAddr + " "))
		w.Write([]byte(r.Header["User-Agent"][0]))
		if r.Header["Cookie"] != nil {
			w.Write([]byte(" " + r.Header["Cookie"][0]))
		}
		w.Write([]byte(opgoweblib.HTML_End))
	} else {
		w.Write([]byte("Пошла нахуй сука китаёзная"))
	}

}
