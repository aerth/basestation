package basestation

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aerth/passwd"
	"github.com/aerth/userinfo"
)

// How to turn []byte to int ?! Use len() !
func hitCounterIncrement(s string) {
	b := userinfo.Read("hits", s)
	b = append(b, byte('0'))
	err = userinfo.Write("hits", s, b)
	if err != nil {
		log.Println(err)
	}

}

func hitCounterRead(s string) int {
	hitCounterIncrement(s)
	i := userinfo.Read("hits", s)
	return len(i)
}

func deleteAllUsers() {
	for id, user := range users {
		passwd.Delete(user.ID)
		delete(users, id)
		userinfo.Delete("user", id)
		users = userinfo.Scan()                       // scan json tables from boltdb
		usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
		usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
	}
}
func deleteObjectBox(id string) {
	box, err := getObjectBox(id)
	if err != nil {
		log.Println(err)
	}
	for _, ob := range box.Objects {
		deleteObject(ob)

	}
	err = userinfo.Delete("objectbox", id)
	if err != nil {
		log.Println(err)
	}
	users = userinfo.Scan()                       // scan json tables from boltdb
	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb

}
func deleteObject(id string) {
	if id == "" {
		return
	}
	deleteFromObjectBox(id)       // remove id string from []objects in box
	userinfo.Delete("object", id) // delete object

	err = os.Remove("files/" + id)
	if err != nil {
		log.Println(err)
	}
	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
}
func deleteFromObjectBox(id string) {
	boxes := userinfo.ScanObjectBoxes()
	for _, box := range boxes {

		var newbox []string
		for _, o := range box.Objects {
			if o != id {
				newbox = append(newbox, o)
			} else {
				log.Printf("Found %s in objectbox: %s, deleting!", id, box.OwnerID)
			}
		}
		box.Objects = newbox
		err = setObjectBox(box)
		if err != nil {
			log.Println(err)
		}
	}
	users = userinfo.Scan()                       // scan json tables from boltdb
	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
}

//
// // appendJson returns old + err if there is an error, otherwise new is old + new. ( not new+old)
// func appendJson(old userinfo.ObjectBox, b []byte) (new userinfo.ObjectBox, err error) {
// 	var ud userinfo.ObjectBox
// 	err = json.Unmarshal(b, &ud)
// 	if err != nil {
// 		return old, err
// 	}
// 	ud.Data = ud.Data + string(b)
//
// 	ud.TimeModified = time.Now()
//
// 	return new, nil
// }
func prepareObject(ownerID string) (ob userinfo.UserObject, err error) {
	ob.TimeCreated = time.Now()
	ob.TimeModified = time.Now()
	ob.OwnerID = ownerID
	ob.ObjectID = generateObjectID()
	ob.Permissions = []string{}
	return ob, err
}
func testFile(b []byte) error {

	return nil
}
func saveObject(ob userinfo.UserObject, b []byte) (int, error) {
	if ob.Filename == "" {
		return 0, errors.New("Blank filename")
	}
	ob.Filename = strings.TrimSuffix(ob.Filename, ".")
	ob.Filename = strings.TrimPrefix(ob.Filename, ".")
	if ob.Extension == "" && strings.Contains(ob.Filename, ".") {
		ob.Extension = strings.Split(ob.Filename, ".")[1]
	}

	file, err := os.Create("files/" + ob.ObjectID)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			return 0, err
		} else {
			os.MkdirAll("files/", 0755)
			file, err = os.Create("files/" + ob.ObjectID)
			if err != nil {
				return 0, err
			}
		}
	}
	n, err := file.Write(b) // save base64 encoded file to disk
	if err != nil {
		return 0, err
	}
	err = file.Close()
	if err != nil {
		return 0, err
	}

	j, err := json.Marshal(ob)
	if err != nil {
		return n, err
	}

	err = userinfo.Write("object", ob.ObjectID, j)
	if err != nil {
		return 0, err
	}
	err = add2box(ob.OwnerID, ob.ObjectID)
	if err != nil {
		return 0, err
	}

	log.Printf("Object %q saved. %v bytes.", ob.ObjectID, n)

	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
	return n, nil
}

func add2box(ownerid string, objectid string) error {
	box, err := getObjectBox(ownerid)
	if err != nil {

		log.Println(err)
		if err.Error() != "new" {
			return err
		}
	}
	log.Println(box)
	time.Sleep(1 * time.Second)
	box.Objects = append(box.Objects, objectid)
	err = setObjectBox(box)
	if err != nil {
		return err
	}
	log.Println(objectid)
	log.Println(box)
	log.Println("Aadded to box")
	return nil
}
func setObjectBox(box userinfo.ObjectBox) error {
	b, err := json.Marshal(box)
	if err != nil {
		return err
	}

	err = userinfo.Write("objectbox", box.OwnerID, b)
	if err != nil {

		return err
	}

	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
	return nil
}

// returns new object + error("new") if doesn't exist (yet)
func getObject(objectid string) (userinfo.UserObject, error) {
	var ob userinfo.UserObject

	b := userinfo.Read("object", objectid)
	if err != nil {
		return ob, err
	}
	if b == nil {

		return ob, errors.New("new")
	}
	//	log.Println(string(b))
	err = json.Unmarshal(b, &ob)
	if err != nil {
		return ob, err
	}
	if *debug {
		log.Println("Object found:", objectid)
	}
	return ob, nil
}

// just a bunch of object IDs grouped by owner
func getObjectBox(userid string) (userinfo.ObjectBox, error) {
	var ud userinfo.ObjectBox
	ud.OwnerID = userid

	//	log.Println("getting objectbox", userid)
	b := userinfo.Read("objectbox", userid)
	//log.Println("GETTING", string(b))
	if err != nil {
		return ud, err
	}
	if b == nil {
		return ud, errors.New("new")
	}
	err = json.Unmarshal(b, &ud)
	if err != nil {
		return ud, err
	}

	return ud, nil
}

//
// func photoInsert(ownerID string, base64data string) (n int, err error) {
// 	var ud userinfo.ObjectBox
// 	ud.OwnerID = ownerID
// 	ud.TimeCreated = time.Now()
// 	ud.TimeModified = time.Now()
//
// 	old, err := userinfoRetrieveData(ownerID)
// 	if err != nil {
// 		if err.Error() != "new" {
// 			log.Println(err)
// 		}
// 	}
// 	if old.Data != "" {
// 		ud.Data = old.Data + " " + base64data
// 	} else {
// 		ud.Data = base64data
// 	}
// 	v, err := json.Marshal(&ud)
// 	err = userinfo.Write("userdata", ownerID, v)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return len(v), nil
// }

// func userinfoInsertData(ownerID string, data string) (n int, err error) {
// 	var ud userinfo.ObjectBox
// 	ud.OwnerID = ownerID
// 	ud.TimeCreated = time.Now()
// 	ud.TimeModified = time.Now()
//
// 	old, err := userinfoRetrieveData(ownerID)
// 	if err != nil {
// 		if err.Error() != "new" {
// 			log.Println(err)
// 		}
// 	}
// 	if old.Data != "" {
// 		ud.Data = old.Data + " " + data
// 	} else {
// 		ud.Data = data
// 	}
// 	v, err := json.Marshal(&ud)
// 	err = userinfo.Write("userdata", ownerID, v)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return len(v), nil
// }

func userinfoInsert(user userinfo.Person) error {
	b, e := json.Marshal(&user)
	if e != nil {
		log.Println(e)
		return e
	}

	err = userinfo.Write("user", user.ID, b)
	if err != nil {
		return err
	}
	users = userinfo.Scan()
	return err
}

// create user pass, returns error if user already exists or can't write to passwd db.
// passwd.Parse() does not need to be called after a signup.
func signup(e string, p []byte) (id string, err error) {

	var person userinfo.Person
	person.FirstName = e
	person.LastName = random(8)
	//person.Age = uint8(ran.Intn(10)) + 18
	//person.Height = uint32(ran.Intn(2))
	person.ZipCode = uint32(ran.Intn(5))
	//person.BodyType = rand8()
	//person.Gender = uint8(ran.Intn(10))
	person.Email = e
	person.ID = generateUserID()

	log.Println(e, "SIGNUP: Going for it!")
	for try := 0; try < 2; try++ {
		plock = true
		err = passwd.Insert(person.ID, p) // insert password
		if err != nil {
			return "", err
		}
		err = passwd.Write() // write the passwd file
		if err != nil {
			return "", err
		}
		log.Println("1")
		plock = false
		//c <- 1
		log.Println("2")
		err = userinfoInsert(person) // write the json bolt db
		if err != nil {
			return "", err
		}

		if err != nil {
			return "", err
		}
		log.Println("Try", try)
		return person.ID, nil
	}
	return person.ID, nil
}

func allObjectBoxes() map[string]userinfo.ObjectBox {
	return userinfo.ScanObjectBoxes()
}
