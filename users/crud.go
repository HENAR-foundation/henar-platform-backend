package users

import (
	"encoding/json"
	"henar-backend/types"
	"log"
	"net/http"

	"golang.org/x/exp/slices"
)

var users []types.UserCredentials // rewrite to db, this is just for speed

func CreateUser(user types.UserCredentialsWithoutId) error {
	if slices.ContainsFunc(users, func(u types.UserCredentials) bool { return u.Email == user.Email }) {
		return nil
	} else {
		users = append(users, types.UserCredentials{
			Id:       int64(len(users)),
			Email:    user.Email,
			Password: user.Password,
		})

		log.Println(users)

	}
	return nil
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Path[len("/api/users/"):]

	if slices.ContainsFunc(users, func(u types.UserCredentials) bool { return u.Email == userId }) {
		user, err := json.Marshal(users[slices.IndexFunc(users, func(u types.UserCredentials) bool { return u.Email == userId })])

		if err != nil {
			http.Error(w, "Error decoding user", http.StatusBadRequest)
		} else {
			w.Write(user)
		}

	} else {
		http.Error(w, "User not found", http.StatusBadRequest)
	}
}
