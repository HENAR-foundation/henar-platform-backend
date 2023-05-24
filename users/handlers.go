package users

// , types.UserCredentials
func CheckEmail(email string) (bool, error) {
	// userIndex := slices.IndexFunc(users, func(u types.UserCredentials) bool { return u.Email == email })
	// if userIndex == -1 {
	// 	return false, types.UserCredentials{}
	// } else {
	// 	user := users[userIndex]
	// 	return true, user
	// }

	// collection, _ := db.GetCollection("users")
	// filter := bson.M{"email": email}
	// var user types.User
	// err := collection.FindOne(context.TODO(), filter).Decode(&user)
	// fmt.Println(err)

	// return true, user
	return true, nil
}
