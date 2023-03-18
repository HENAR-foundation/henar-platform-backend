package auth

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("vsljvnfsljbnfsljbnfblkjnf")) // remove session key

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session.id")

		if session.Values["authentificated"] != nil && session.Values["authentificated"] != false {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}

	})
}

func WithUserCreds(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")

	if session.Values["authentificated"] != nil && session.Values["authentificated"] != false {
		w.Write([]byte(time.Now().String()))
	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}

// func SessionMiddleware(h http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		session, _ := store.Get(r, "session.id")

// 		if session.Values["authentificated"] != nil && session.Values["authentificated"] != false {
// 			w.Write([]byte(time.Now().String()))
// 		} else {
// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 		}

// 		h(w, r)
// 	}
// }
