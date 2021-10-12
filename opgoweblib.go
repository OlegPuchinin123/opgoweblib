package opgoweblib

import (
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const VERSION = "1.0"

func Version() string {
	return VERSION
}
func Check_cookie(cookie_name string, r *http.Request) string {
	var (
		cookie string
		idx    int
		spl    []string
		s      string
	)

	if r.Header["Cookie"] != nil {
		cookie = r.Header["Cookie"][0]
		spl = strings.Split(cookie, "; ")
		for _, s = range spl {
			if strings.HasPrefix(s, cookie_name+"=") {
				idx = strings.Index(s, "=")
				cookie = s[idx+1:]
				break
			}
		}
	}
	return cookie
}
