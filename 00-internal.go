/*
Basestation is a web application

Copyright (c) 2016 aerth

GPL3 License
*/

package basestation

import (
	"flag"
	"net"
	"net/http"

	"github.com/aerth/userinfo"
	"github.com/gorilla/sessions"
)

//
// type personType struct {
// 	ID, FirstName, LastName, NickName, Email, Data string `json:",omitempty"`
// 	Gender, BodyType, Age                          uint8  `json:",string,omitempty"`
// 	Height, ZipCode                                uint32 `json:",string,omitempty"`
// }

type sessionView struct {
	BaseURI   string
	Extension string
	Folder    string
	Name      string
	Caching   bool
	Vars      map[string]interface{}
	request   *http.Request
}

var (
	store       = sessions.NewCookieStore([]byte("1234567812345678123456781234567812345678123456781234567812345678"))
	antiCSRFkey = []byte("1234567812345678123456781234567812345678123456781234567812345678")
	sessionName = "nameforcookie"
	listener    net.Listener
	route       http.Handler
	//visitorLog  *log.Logger
	err              error
	oglistener       net.Listener
	plist            []string                       // passwd users
	users            map[string]userinfo.Person     // json userinfo users
	usersObjects     map[string]userinfo.UserObject // json userinfo users
	usersObjectBoxes map[string]userinfo.ObjectBox  // json userinfo users
	// flags
	version = flag.Bool("v", false, "Display version information and exit")
	// server
	port       = flag.String("port", "8090", "Port to listen on")
	bind       = flag.String("bind", "0.0.0.0", "Default: 0.0.0.0 (all interfaces)...\n\tTry -bind=127.0.0.1")
	dbname     = flag.String("dbname", ".db", "db name")
	passwdname = flag.String("passwd", ".passwd", "passwd file name")
	// logging
	debug       = flag.Bool("debug", false, "More verbose output. For figuring out what went wrong where.")
	templateDir = "./templates"
)
