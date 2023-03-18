package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("vsljvnfsljbnfsljbnfblkjnf")) // remove session key

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if r.URL.Path != "/open-urls" {
		// 	next.ServeHTTP(w, r)
		// } else {
		session, _ := store.Get(r, "session.id")

		if session.Values["authentificated"] != nil && session.Values["authentificated"] != false {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
		// }

	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

}
