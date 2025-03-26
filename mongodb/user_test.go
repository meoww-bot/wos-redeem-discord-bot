package mongodb

// func TestGetUser(t *testing.T) {
// 	user, err := GetUser(int64(user_id))

// 	if err != nil {
// 		t.Log(err)
// 	}

// 	t.Logf("%+v", user)
// 	if !(assert.Equal(t, user.UID, int64(user_id), "they should be equal")) {
// 		t.Fail()
// 	}

// }

// func TestListAllUser(t *testing.T) {
// 	users, err := GetAllUser()

// 	if err != nil {
// 		t.Log(err)
// 	}

// 	var responseString string = fmt.Sprintf("UserID - Nickname (%d users)", len(users))

// 	for _, user := range users {

// 		responseString += fmt.Sprintf("\n%d - %s", user.FID, user.Nickname)
// 	}

// 	t.Log(len(responseString))

// }

// func TestGetAllUser(t *testing.T) {
// 	users, err := GetAllUser()

// 	if err != nil {
// 		t.Log(err)
// 	}

// 	t.Logf("%+v", users)

// 	t.Log(len(users))
// }

// func TestGetUserRealStoveLv(t *testing.T) {
// 	user, err := GetUser(417561)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// t.Logf("StoveLv的类型: %T\n", user.StoveLv)

// 	t.Log(user.GetUserRealStoveLv())
// }
