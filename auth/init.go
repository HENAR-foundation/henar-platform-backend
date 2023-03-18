package auth

import (
	"encoding/json"
	"log"
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

func AdminMiddleware(next http.Handler) http.Handler {
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

type Credentials struct {
	Email    string
	Password string
}

var users = []Credentials{
	{Email: "test", Password: "pass"},
	{Email: "test1", Password: "word"},
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Fatal(err)
	}

	for _, user := range users {
		if user.Email == creds.Email && user.Password == creds.Password {
			session, _ := store.Get(r, "session.id")

			session.Values["authentificated"] = true
			session.Save(r, w)
			w.Write([]byte("Login"))
			break
		} else {
			continue
		}
	}

	log.Println(creds.Email)
	log.Println(creds.Password)

	http.Error(w, "Email or password incorrent", http.StatusBadRequest)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	session.Values["authenticated"] = false
	session.Save(r, w)
	w.Write([]byte("Logout"))
}
