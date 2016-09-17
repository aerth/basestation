package basestation

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/websocket"

	"github.com/aerth/userinfo"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func setroute() *mux.Router {

	r := mux.NewRouter()

	//Redirect 404 errors home?
	r.NotFoundHandler = http.HandlerFunc(four04Process)

	// Home Page
	r.HandleFunc("/", srvHome)
	r.HandleFunc("/dashboard", srvDashboard)
	r.HandleFunc("/signup", srvSignup)
	r.HandleFunc("/forum/{whatever}", srvForum)
	r.HandleFunc("/forum/{whatever}/{whatever}", srvForum)
	r.HandleFunc("/forum", srvForum)
	r.HandleFunc("/login", srvLogin)
	r.HandleFunc("/logout", srvLogout)
	r.HandleFunc("/halt", srvShutdown)
	//r.HandleFunc("/echo", http.HandlerFunc(ExampleHandler))
	r.HandleFunc("/testing", srvTesting)
	r.HandleFunc("/file/{whatever}", srvFiles)

	r.HandleFunc("/dummy", srvDummy)

	http.Handle("/", r)

	return r
}

func srvFiles(w http.ResponseWriter, r *http.Request) {
	sess, user := roadblock(w, r)
	myID := user.ID
	// u, err := url.ParseRequestURI(r.RequestURI)
	// if err != nil {
	// 	log.Println(err)
	// }
	//path := u.EscapedPath()
	path := r.URL.Path[len("/file/"):]

	pathparts := strings.Split(path, ".")

	if len(pathparts) == 1 {
		objectid := pathparts[0]

		log.Println("pathparts[0]", objectid)
		getext, err := getObject(objectid)
		if err != nil { // err.Error = "new" if no object by that ID exists
			log.Println(err)
			sess.AddFlash("No object by that ID.")
			sess.Save(r, w)
			redirect(w, r, "/testing")
			return
		}
		if getext.Extension != "" {
			log.Println("Correcting extension")
			redirect(w, r, "/file/"+getext.ObjectID+"."+getext.Extension)
			return
		}
	}

	// This is a true file. ObjectID + "." + extension
	var objectid string
	var extension string
	if len(pathparts) == 2 {
		extension = pathparts[1]
		objectid = pathparts[0]
		//fileid = strings.TrimSuffix(nameparts[0], "."+extension)
	}

	object, err := getObject(pathparts[0])
	if err != nil {
		log.Println("No Object by that objectID")
		redirect(w, r, "/testing")
		return
	}

	if object.Extension != extension {
		log.Println("Correcting extension Lvl. 2")
		redirect(w, r, "/file/"+object.ObjectID+"."+object.Extension)
		return
	}

	log.Printf("Request ObjectID: %s #%v\n", objectid, hitCounterRead(objectid))

	r.ParseForm()
	d := r.FormValue("delete")

	if d == "true" && object.OwnerID == myID {
		log.Println("Call for delete")
		deleteObject(objectid)
		deleteFromObjectBox(objectid)
		deleteObjectFromFilesystem(objectid)
		redirect(w, r, "/testing")
		return
	}
	unshare := r.FormValue("unshare")

	if unshare == "true" && object.OwnerID == myID {
		log.Printf("Call for delete of %s by %s\n", myID, object.ObjectID)
		deleteObject(objectid)
		deleteFromObjectBox(objectid)
		deleteObjectFromFilesystem(objectid)
		redirect(w, r, "/testing")
		return
	}
	var permindex int
	var granted bool
	if object.IsPublic == false && object.OwnerID != myID {
		in := sort.SearchStrings(object.Permissions, myID)
		log.Println(in, object.Permissions)

		if len(object.Permissions) > 1 && in == len(object.Permissions) {
			log.Println("No permission!!!")
			sess.AddFlash("No permission for object by that ID.")
			sess.Save(r, w)
			redirect(w, r, "/testing")
			return

		}

		for i, ids := range object.Permissions {
			if granted != true {
				if ids == myID {
					permindex = i
					granted = true
				}

			}
		}
		if !granted {
			log.Println("No permission!")
			sess.AddFlash("No permission for object by that ID.")
			sess.Save(r, w)
			redirect(w, r, "/testing")
			return
		}
	}

	if objectid == "" {
		log.Println("Blank objectID")
		redirect(w, r, "/testing")
		return
	}

	file, err := os.Open("files/" + objectid)
	if err != nil {
		log.Println(err)
		redirect(w, r, "/testing")
		return
	}
	info, err := file.Stat()
	if err != nil {
		log.Println(err)
		redirect(w, r, "/testing")
	}

	var b = make([]byte, info.Size()+100)
	n, err := file.Read(b)
	if err != nil {
		log.Println(err)
		redirect(w, r, "/testing")
	}
	var out = make([]byte, len(b)+100)
	g, err := base64.StdEncoding.Decode(out, b[:n])

	switch extension {
	case "png":
		w.Header().Set("Content-Type", "image/png")

	case "jpg", "jpeg":
		w.Header().Set("Content-Type", "image/jpeg")

	case "gif":
		w.Header().Set("Content-Type", "image/gif")
	case "zip":
		w.Header().Set("Content-Type", "application/zip")
	case "php":
		w.Header().Set("Content-Type", "text/plain")
	case "exe", "virus":
		w.Write([]byte("This file is under review."))
		return
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
	var pubstate string
	if object.IsPublic == true {
		pubstate = "public"
	} else {
		pubstate = "private"
	}
	log.Printf("serving file:\n\tfile\n\t%s\n\towner:%s \n\tto:   %s (%v) (%s)\n\n", object.ObjectID, object.OwnerID, sess.Values["UID"].(string), permindex, pubstate)
	w.Write(out[:g])

}
func four04Process(w http.ResponseWriter, r *http.Request) {
	reHome(w, r)
}

func srvHome(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, sessionName)
	v := newView(r)
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = flashes
	}
	if sess.Values["ACTIVE"] == true {
		redirect(w, r, "dashboard")
	}

	data := map[string]interface{}{
		"MyFlashes": v.Vars["flashes"],
		"Vars":      v.Vars,

		csrf.TemplateTag: csrf.TemplateField(r),
	}

	triplecombo(w, "home", data)
}
func reHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}
func reLogin(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess, _ := store.Get(r, sessionName)
	log.Println("logging out session values:", sess.Values)
	// If user is authenticated
	if sess.Values["UID"] != "" || sess.Values["ACTIVE"] != "" {
		logout(sess)
		log.Println("cleared it:", sess.Values)

		sess.Save(r, w)
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func redirect(w http.ResponseWriter, r *http.Request, s string) {
	http.Redirect(w, r, s, http.StatusFound)
}

func getsesh(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	sess, _ := store.Get(r, sessionName)

	if sess.Values["UID"] == nil {
		return sess, true // error means user is not logged in, and this is a new session.
	}

	return sess, false
}

func srvDashboard(w http.ResponseWriter, r *http.Request) {
	sess, _ := getsesh(w, r)

	if sess.Values == nil {
		log.Println("nil sess")
		reLogin(w, r)
		return
	}

	if sess.Values["UID"] == nil {
		log.Println("nil sess uid")
		reLogin(w, r)
		return
	}

	myUser, err := getUserDataByID(sess.Values["UID"].(string))

	if err != nil {
		log.Println(err)
		reLogin(w, r)
		return
	}

	if myUser.ID != sess.Values["UID"].(string) {
		log.Println("no user, bad session")
		reLogin(w, r)
		return
	}
	log.Println(myUser.ID, "ENTER")
	// if new == true {
	// 	log.Println("new session going to login!")
	// 	redirect(w, r, "login")
	// 	return
	// }
	v := newView(r)
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = flashes
	}
	sess.Save(r, w)

	myUserID := myUser.ID
	myuserdata := myUser
	data := map[string]interface{}{
		"MyFlashes":      v.Vars["flashes"],
		"Vars":           v.Vars,
		"AllUsers":       users,
		"AllUsersPasswd": plist,
		"MyUser":         myUserID,
		"MyUserData":     myuserdata,

		csrf.TemplateTag: csrf.TemplateField(r),
	}

	triplecombo(w, "dashboard", data)
}
func srvBlank(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("BLANK"))
}

func srvSignup(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.RequestURI, r.Cookies())
	sess, _ := store.Get(r, sessionName)
	log.Println(r.Header)
	lol := make([]byte, 1024*1000)
	n, err := r.Body.Read(lol)
	if err != nil {
		if err.Error() != "EOF" {
			log.Println(err)
		}
	}
	log.Println(string(lol[:n]))
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("SIGNUP"))
	switch r.Method {
	case "POST":
		e := r.ParseForm()
		if e != nil {
			log.Println(e)
			redirect(w, r, "/signup")
			return
		}
		u := r.FormValue("u")
		p := []byte(r.FormValue("p"))
		if u == "" || p == nil {
			log.Println("blank user")
			redirect(w, r, "/signup")
			return
		}
		// inserting into passwd file
		id, e := signup(u, p)
		if e != nil {
			log.Println(e)
			redirect(w, r, "/signup")
			return
		}
		log.Println("New User:", u, id)
		sess.Values["UID"] = id
		sess.Save(r, w)
		reHome(w, r)

	case "GET":
		v := newView(r)
		if flashes := sess.Flashes(); len(flashes) > 0 {
			v.Vars["flashes"] = flashes
		}
		sess.Save(r, w)
		data := map[string]interface{}{
			"MyFlashes":      v.Vars["flashes"],
			"Vars":           v.Vars,
			csrf.TemplateTag: csrf.TemplateField(r),
			//	"CaptchaId":      captcha.NewLen(CaptchaLength + rand.Intn(CaptchaVariation)),
		}

		triplecombo(w, "signup", data)

	}

}

func srvLogout(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess, _ := store.Get(r, sessionName)
	log.Println("logging out session values:", sess.Values)
	// If user is authenticated
	if sess.Values["UID"] != "" || sess.Values["ACTIVE"] != "" {
		logout(sess)
		log.Println("cleared it:", sess.Values)

		sess.Save(r, w)
	}

	http.Redirect(w, r, "/", http.StatusFound)

}

func srvLogin(w http.ResponseWriter, r *http.Request) {

	sess, _ := store.Get(r, sessionName)
	v := newView(r)
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = flashes
	}
	if sess.Values["ACTIVE"] == true {
		redirect(w, r, "dashboard")
	}
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte("SIGNUP"))
	switch r.Method {
	case "POST":
		e := r.ParseForm()
		if e != nil {
			log.Println(e)
			redirect(w, r, "/login")
			return
		}
		email := r.FormValue("u")
		p := []byte(r.FormValue("p"))
		if email == "" || p == nil {
			log.Println("blank user")
			redirect(w, r, "/login")
			return
		}

		u, _, err := getUserDataByEmail(email)
		if err != nil {
			log.Println("no user by that email", u.ID, email)
			redirect(w, r, "login")
			return
		}
		// inserting into passwd file
		switch loginCheck(u.ID, p) { // bubba the bouncer
		case true:

			log.Printf("%q logged in!\n", u.ID)
			// Set some session values.
			sess.Values["UID"] = u.ID
			sess.Values["ACTIVE"] = true

			// Save it before we write to the response/return from the handler.
			sess.AddFlash("Hello, welcome back!")
			log.Println(time.Now(), "User", u.ID, "just logged in.")
			sess.Save(r, w)

			redirect(w, r, "/dashboard")
			return
		default:
			log.Printf("%q wrong password!\n", u.ID)
		}

		reHome(w, r)

	case "GET":

		data := map[string]interface{}{

			"MyFlashes":      v.Vars["flashes"],
			"Vars":           v.Vars,
			csrf.TemplateTag: csrf.TemplateField(r),
			//	"CaptchaId":      captcha.NewLen(CaptchaLength + rand.Intn(CaptchaVariation)),
		}

		triplecombo(w, "login", data)

	}

}
func srvShutdown(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}
func srvDummy(w http.ResponseWriter, r *http.Request) {

	id, err := dummyUserJSON()
	if err != nil {
		log.Println(err)
		redirect(w, r, "/dashboard")
		return
	} else {
		log.Println("Got new dummy:", id)
		user, err := getUserDataByID(id)
		if err != nil {
			log.Println(err)
			redirect(w, r, "/dashboard")
			return
		}

		log.Println("New Dummy:", id, user.Email)
	}
	redirect(w, r, "/dashboard")
}

var c = make(chan int)

var plock bool

// Echo the data received on the WebSocket.
func srvWebsocketEcho(ws *websocket.Conn) {
	var b = make([]byte, 1024)
	n, err := ws.Read(b)
	if err != nil {
		log.Println(err)
		return
	}
	ws.Write(b[:n-2])
	e := ws.Close()
	if e != nil {
		log.Println(err)
	}

}

// This example demonstrates a trivial echo server.
func srvWebsocketEchoLaunch() {
	http.Handle("/echo", websocket.Handler(srvWebsocketEcho))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func srvUpdateProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Updating profile")
	sess, _ := getsesh(w, r)

	if sess.Values == nil {
		log.Println("nil sess")
		reLogin(w, r)
		return
	}

	if sess.Values["UID"] == nil {
		log.Println("nil sess uid")
		reLogin(w, r)
		return
	}
	log.Println("running srvUpdateData")
	myUser, err := getUserDataByID(sess.Values["UID"].(string))
	if err != nil {
		log.Println(err)
		reLogin(w, r)
		return
	}
	log.Println("Post detected.")

	// nickname
	if r.FormValue("nickname") != myUser.NickName {
		if name2id(r.FormValue("nickname")) == "0" {
			myUser.NickName = r.FormValue("nickname")
		} else {
			sess.AddFlash("@username exists already")
			sess.Save(r, w)
		}
	}

	// email
	if r.FormValue("email") != myUser.Email && r.FormValue("email") != "" {
		if email2id(r.FormValue("email")) == "0" {
			myUser.Email = r.FormValue("email")
		} else {
			sess.AddFlash(r.FormValue("email") + " exists already")
			sess.Save(r, w)
		}
	}

	// gender
	log.Println(r.FormValue("gender"), myUser.Gender)
	if r.FormValue("gender") != string(myUser.Gender) { // gender change?
		if i := gender2int(r.FormValue("gender")); i != 0 {
			myUser.Gender = i

		} else {
			log.Println(r.FormValue("gender"), i)
			sess.AddFlash("invalid gender")
			sess.Save(r, w)
		}
	}

	// write changes
	er := updateUserData(myUser)
	if er != nil {
		sess.AddFlash(er)
		sess.Save(r, w)
		log.Println(er)
	}
	log.Println("made it down")
	redirect(w, r, "/testing")
}
func srvUpdatePassword(w http.ResponseWriter, r *http.Request) {
	// passwd
	sess, user := roadblock(w, r)
	if user.ID == "" {
		reLogin(w, r)
		return
	}

	if r.FormValue("p1") == r.FormValue("p2") && r.FormValue("p2") != "" {
		err = updateUserPasswd(user.ID, []byte(r.FormValue("p2")))
		if err != nil {
			log.Println(err)
		}
	} else {
		sess.AddFlash("new passwords dont match" + r.FormValue("p1") + r.FormValue("p2"))
		sess.Save(r, w)
	}
	log.Println("made it down")
	redirect(w, r, "/testing")
}

func srvAddPhoto(w http.ResponseWriter, r *http.Request) {
	log.Println("running srvAddPhoto")
	sess, _ := getsesh(w, r)

	if sess.Values == nil {
		log.Println("nil sess")
		reLogin(w, r)
		return
	}

	if sess.Values["UID"] == nil {
		log.Println("nil sess uid")
		reLogin(w, r)
		return
	}

	myUser, err := getUserDataByID(sess.Values["UID"].(string))
	if err != nil {
		log.Println(err)
		reLogin(w, r)
		return
	}
	log.Println("Post detected.")

	pstring := r.FormValue("public")
	var public bool
	if pstring == "off" {
		public = false
	}
	if pstring == "on" {
		public = true
	}

	perms := cleanFormPermissions(myUser.ID, r.FormValue("permissions")) // can be "", ""

	file, fileheader, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
	}

	var b = make([]byte, 1024*2000)
	n, err := file.Read(b)
	if err != nil {
		log.Println(err)
	}
	out := base64.StdEncoding.EncodeToString(b[:n])

	log.Println("File Received and Base64 Encoded:", len(out), "bytes")

	ob, err := prepareObject(myUser.ID)
	if err != nil {
		log.Println(err)
	}
	ob.Filename = fileheader.Filename
	ob.Size = n
	ob.Permissions = perms
	log.Println("public?", public, r.FormValue("public"))
	ob.IsPublic = public
	log.Println(ob)
	_, err = saveObject(ob, []byte(out))
	if err != nil {
		log.Println(err)
	}
	sess.AddFlash("New object: " + ob.ObjectID)
	sess.Save(r, w)
	redirect(w, r, "/testing")
	return
}

func cleanFormPermissions(u, s string) []string {
	var perms = []string{u} // decode permissions
	r := strings.Split(s, ",")
	for _, a := range r {
		a = strings.TrimSpace(a)
		a = html.EscapeString(a)
		id := name2id(a)
		if id == "0" {
			log.Println(err)
			continue

		}
		log.Println("Found name2id:", s, a, id)
		perms = append(perms, id)
	}

	return perms
}
func srvTesting(w http.ResponseWriter, r *http.Request) {
	sess, myUser := roadblock(w, r)
	log.Println(myUser.ID, "logged in")
	var needed []string

	if r.Method == "POST" {
		switch r.FormValue("action") {
		case "":
			redirect(w, r, "/testing")
			return
		case "addphoto":
			srvAddPhoto(w, r)
			return
		case "profile":
			srvUpdateProfile(w, r)
			return
		case "password":
			srvUpdatePassword(w, r)
			return
		case "deletebox":
			log.Println("deleting dataz:", myUser.ID)
			deleteObjectBox(myUser.ID)
			redirect(w, r, "/testing")
			return
		default:
			redirect(w, r, "/testing")
			return

		}

	}
	if myUser.NickName == "" {
		needed = append(needed, "NickName")
	}
	if myUser.Gender == 0 {
		needed = append(needed, "Gender")
	}
	if len(needed) > 0 {
		bangUpdateProfile(w, r, needed)
		return
	}
	v := newView(r)
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = flashes
	}
	sess.Save(r, w)
	myObjectBox, err := getObjectBox(myUser.ID)
	if err != nil {
		if err.Error() == "new" {
			log.Println("Creating Brand new objectbox")
			if err := setObjectBox(myObjectBox); err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	}
	var mapMyObjects = map[string]userinfo.UserObject{}
	for _, k := range myObjectBox.Objects {
		o, err := getObject(k)
		if err != nil {
			log.Println(err)
		} else {
			mapMyObjects[k] = o
		}
	}
	users = userinfo.Scan()                       // scan json tables from boltdb
	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
	myUserID := myUser.ID
	myuserdata := myUser
	shared := sharedWithUser(myUserID)

	data := map[string]interface{}{
		"MyFlashes":      v.Vars["flashes"],
		"Vars":           v.Vars,
		"AllUsers":       users,
		"AllUsersPasswd": plist,
		"SharedWithMe":   shared,
		"MyUser":         myUserID,
		"MyUserData":     myuserdata,
		"MyObjectBox":    myObjectBox,
		"MyObjects":      mapMyObjects,
		"AllUserDataBox": allObjectBoxes(),

		csrf.TemplateTag: csrf.TemplateField(r),
	}

	triplecombo(w, "testing", data)
}
func removeEmpty(s []string) []string {
	var d []string
	for _, v := range s {
		if v != "" {
			d = append(d, v)
		}
	}
	return d
}

func deleteObjectFromFilesystem(objectid string) error {

	err := os.Remove("files/" + objectid)
	if err != nil {
		return err
	}
	return nil
}
func checksesh(sess *sessions.Session) error {
	if sess.Values == nil {
		return errors.New("Nil session")
	}

	if sess.Values["UID"] == nil {
		return errors.New("Nil session")

	}
	return nil
}
func roadblock(w http.ResponseWriter, r *http.Request) (*sessions.Session, userinfo.Person) {
	sess, _ := getsesh(w, r)

	err = checksesh(sess)
	if err != nil {
		log.Println(err)
		return nil, userinfo.Person{}
	}

	uid := sess.Values["UID"].(string)
	user, err := getUserDataByID(uid)
	if err != nil {
		return nil, userinfo.Person{}
	}

	return sess, user
}

func checkprofile(u userinfo.Person) []string {
	var g []string
	if u.NickName == "" {
		g = append(g, "NickName")
	}
	if u.Gender == 0 {
		g = append(g, "Gender")
	}
	if u.Height == 0 {
		g = append(g, "Height")
	}
	if u.ZipCode == 0 {
		g = append(g, "ZipCode")
	}
	if u.BodyType == 0 {
		g = append(g, "BodyType")
	}
	if u.LookingFor == 0 {
		g = append(g, "LookingFor")
	}
	return g
}

func bangUpdateProfile(w http.ResponseWriter, r *http.Request, g []string) {
	sess, user := roadblock(w, r)
	log.Println("Roadblock:", user.ID, g)
	if user.ID == "" {
		reLogin(w, r)
		return
	}
	if len(g) == 0 {
		redirect(w, r, "/dashboard?allgood")
	}
	v := newView(r)
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = flashes
	}
	sess.Save(r, w)
	data := map[string]interface{}{
		"MyFlashes":           v.Vars["flashes"],
		"Vars":                v.Vars,
		"NeededProfileFields": g,

		csrf.TemplateTag: csrf.TemplateField(r),
	}

	triplecombo(w, "bangprofile", data)
	sess.Save(r, w)
}
func gender2int(s string) uint8 {
	s = strings.Title(s)
	switch s {
	case "1", "female":
		return 1
	case "2", "male":
		return 2
	case "3", "couple":
		return 3
	case "4", "transgender":
		return 4
	case "5", "other":
		return 5
	case "0":
		return 0
	default:
		return 0
	}

}

func validateExtension(s string) (string, error) {
	x := getExtension(s)
	switch x {
	// List valid Extensions
	case "png", "jpeg", "jpg", "gif", "mp4", "avi", "mpeg", "m4a", "mp3", "wav",
		"tgz", "tar.gz", "tar.xz", "bz2.gz",
		"pdf", "txt", "rtf", "doc", "zip", "pgp", "gpg", "asc", "torrent", "db", "backup", "html", "php", "go":
		return s, nil
	}
	return x, errors.New("Invalid extension")
}

func getExtension(s string) string {
	s = strings.TrimPrefix(s, ".")
	s = strings.TrimSuffix(s, ".")
	if strings.HasSuffix(s, ".tar.gz") {
		return "tar.gz"
	}
	if strings.HasSuffix(s, ".tar.xz") {
		return "tar.xz"
	}
	if strings.HasSuffix(s, ".bz2.gz") {
		return "bz2.gz"
	}
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return ""
	}

	if len(parts) == 2 {
		return parts[1]
	}

	return parts[len(parts)-2]
}
