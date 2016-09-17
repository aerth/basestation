package basestation

import (
	"errors"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/aerth/passwd"
	"github.com/aerth/userinfo"
	"github.com/gorilla/sessions"
)

// Random
var ran *rand.Rand

func init() {
	ran = rand.New(rand.NewSource(time.Now().UnixNano())) // new random source
}

// loginCheck checks user pass, returns true if matches the passwd file.
func loginCheck(u string, p []byte) bool {
	if passwd.Match(u, p) {
		return true
	}
	return false
}

// Logout step 1 killsession deletes all the current session values
func logout(sess *sessions.Session) {
	if sess == nil {
		return
	}
	// Clear out all stored values in the cookie
	for k := range sess.Values {
		delete(sess.Values, k)
	}
}

func getUserDataByID(s string) (p userinfo.Person, e error) {
	if users == nil {
		return p, errors.New("No user database")
	}
	if users[s].ID == "" {
		return p, errors.New("No user")
	}
	return users[s], nil
}

// getUserDataByEmail very important func, id string is for convenience.
func getUserDataByEmail(s string) (p userinfo.Person, userid string, e error) {
	if users == nil {
		return p, "", errors.New("No user database")
	}

	for userid, user := range users { // This could take a while if there is a big usertable? need benchmark
		log.Printf("Scan by email: is %q == %q ?\n", s, user.Email)
		if user.Email == s && userid == user.ID {
			log.Println("Got a user by email!")
			return user, userid, nil
		}
	}

	return p, "", errors.New("No user")
}

// Needs correct ID in p
func updateUserData(p userinfo.Person) error {
	err = userinfoInsert(p)
	if err != nil {
		return err
	}
	return nil
}

func updateUserPasswd(id string, password []byte) error {
	err = passwd.Update(id, password)
	if err != nil {
		return err
	}
	return nil
}

// random character
func random(n int) string {
	runes := []rune("abcdefg1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return strings.TrimSpace(string(b))
}

//rand8 returns a random uint8 (0-255)
func rand8() uint8 {
	random32 := ran.Uint32()
	return uint8(random32)
}

// generateUserID() creates a new unique ID for use as a key in the key[value] stores
// This way, each user has a unique internal identifier that can't change.
// This gets used in the passwd file and the userinfo bolt bucket
func generateUserID() string {
	r := random(32)
	try := userinfo.Read("user", r)
	if try == nil {
		return r
	}
	if string(try) != "" {
		if *debug {
			log.Println("Collision:", r)
		}
		return generateUserID() // almost NEVER gets called so far.
	}
	if *debug {
		log.Println("Generated User ID:", r)
	}
	return r

}

func generateObjectID() string {
	r := random(64)
	try := userinfo.Read("objects", r)
	if try == nil {
		return r
	}
	if string(try) != "" {
		if *debug {
			log.Println("Collision:", r)
		}
		return generateObjectID() // almost NEVER gets called so far.
	}
	if *debug {
		log.Println("Generated Object ID:", r)
	}
	return r

}

func sharedWithUser(userid string) map[string]userinfo.UserObject {
	var shared = map[string]userinfo.UserObject{}
	usersObjects = userinfo.ScanObjects() // scan json tables from boltdb

	for _, ob := range usersObjects {
		p := ob.Permissions
		for _, perm := range p {
			if perm == userid && ob.OwnerID != userid {
				shared[ob.ObjectID] = ob
			}
		}
	}
	return shared

}
