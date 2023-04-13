package users

import (
	"henar-backend/types"

	"golang.org/x/exp/slices"
)

func CheckEmail(email string) (bool, types.UserCredentials) {
	userIndex := slices.IndexFunc(users, func(u types.UserCredentials) bool { return u.Email == email })
	if userIndex == -1 {
		return false, types.UserCredentials{}
	} else {
		user := users[userIndex]
		return true, user
	}
}
