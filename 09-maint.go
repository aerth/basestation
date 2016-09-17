package basestation

import (
	"log"
	"sort"

	"github.com/aerth/passwd"
	"github.com/aerth/userinfo"
)

// cleanUserDatabaseWithoutMercy syncs the userinfo and passwd databases with no mercy
func cleanUserDatabaseWithoutMercy() {
	plist = passwd.List()
	users = userinfo.Scan()
	sort.Strings(plist)
	// range the json table, check passwd table for each one.
	var i int
	for id, user := range users {
		//	log.Println("Ranging:", id, user.ID, user.Email)
		if user.ID != id {
			//log.Println("Unequal ID! deleting!")
			err = userinfo.Delete("user", user.ID)
			if err != nil {
				log.Println(err)
			}
			err = userinfo.Delete("user", id)
			if err != nil {
				log.Println(err)
			}
			passwd.Delete(user.ID)
			passwd.Delete(id)
			continue
		}
		index := sort.SearchStrings(plist, user.ID)

		if index > len(plist)-1 {
			log.Println("json user:", user.ID, "is not in range passwd users, deleting from json!")
			err = userinfo.Delete("user", user.ID)
			if err != nil {
				log.Println(err)
			}

		} else {
			//log.Println("json user:", user.ID, index)
			i++
		}
	}
	log.Printf("%v users in boltDB", i)
	i = 0
	// range the passwd table, check json table for each one.
	for _, userid := range plist {
		_, err = getUserDataByID(userid)
		if err != nil {
			log.Println("passwd user :", userid, "is not in range json users, deleting from passwd!")
			passwd.Delete(userid)
		} else {
			i++
		}
	}
	log.Printf("%v users in passwd file", i)
	err = passwd.Write()
	if err != nil {
		log.Println(err)
	}

}

func cleanObjects() {

	// Check OwnerID
	log.Println("Cleaning Objects:", len(usersObjects))
	for _, ob := range usersObjects {
		_, err = getUserDataByID(ob.OwnerID)
		if *debug {
			log.Println("Checking Object:", ob.ObjectID)
		}
		if err != nil {
			log.Println(err)
			log.Println("Orphan Object's owner disappeared.")
			deleteObject(ob.ObjectID)
			continue
		}

		ownerbox, err := getObjectBox(ob.OwnerID)
		if err != nil {
			if err.Error() != "new" {
				log.Println(err)
			}
			log.Println("Orphan Object's owner has no box")
			deleteObject(ob.ObjectID)
			continue
		}

		for _, oid := range ownerbox.Objects {
			_, err = getObject(oid)
			if err != nil {
				if err.Error() != "new" {
					log.Println(err)
				}
				log.Println("Orphan Object doesn't exist in owners box.")
				deleteObject(oid)
				continue
			}
		}

	}

	// Check Owner's ObjectBox
	for _, ob := range usersObjects {
		_, err := getUserDataByID(ob.OwnerID)
		if err != nil {
			log.Println(err)
			log.Println("^ means Orphan Object")
			continue
		}
	}
}
