// OK to delete this file

package basestation

//
// func allusers() []string {
// 	return passwd.List()
// }
//
// func allusers2() map[string]userinfo.Person {
// 	return users
// }
//
// func printUserTable() {
// 	for i, user := range users {
// 		fmt.Println(i, user.ID, user.Age, user.FirstName, user.LastName)
// 	}
// }

// generate user
func dummyUserJSON() (id string, err error) {

	email := "blowtorch" + random(8) + "@gmail.com"

	id, err = signup(email, []byte("password"))
	return id, err

}

//
// // always returns a user.
// func user(id string) userinfo.Person {
// 	//var p userinfo.Person
// 	return users[id]
// }
