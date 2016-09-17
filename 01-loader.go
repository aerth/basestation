package basestation

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"time"

	forum "github.com/aerth/fforum"
	"github.com/aerth/passwd"
	"github.com/aerth/userinfo"
	"github.com/gorilla/csrf"
	"github.com/hydrogen18/stoppableListener"
)

func init() {
	// Verbose logging with file name and line number
	log.SetFlags(log.Lshortfile + log.Ldate + log.Ltime)
	//log.SetFlags(log.Ldate + log.Ltime)
	runtime.GOMAXPROCS(runtime.NumCPU())
	//runtime.Breakpoint()
	rand.Seed(time.Now().UnixNano())

	//visitorLog = log.New(nil, "", log.LstdFlags)
}
func Init() {
	passwd.SetLocation(*passwdname)

	// defer func() {
	// 	log.Println("Saving passwd db...")
	// 	e := passwd.Write()
	// 	if e != nil {
	// 		log.Println(e)
	// 	}
	//
	// }()

	flag.Parse()
	// parse -flags
	route = setroute()
	// route url to functions
	passwd.Parse() // read passwd file, create usertable, enable passwd.Match for logins
	userinfo.Init(*dbname, []string{"user", "objectbox", "object"})
	forum.Init("forum.db")

	//
	log.Println("All Replies")
	ar100 := forum.AllReplies()
	//
	for _, k := range ar100 {
		log.Printf("ID\t TO\t BODY")
		log.Printf("%q\t %q\t %q", k.ID, k.To, k.Body)
	}
	//
	log.Println("Replies to topic: 9c15660d1824a13b8194fa0222g2176b7520faf5003dg7039c88f42af377a495")
	ar200 := forum.AllRepliesOf("9c15660d1824a13b8194fa0222g2176b7520faf5003dg7039c88f42af377a495")
	//
	for _, k := range ar200 {
		log.Printf("ID\t TO\t BODY")
		log.Printf("%q\t %q\t %q", k.ID, k.To, k.Body)
	}

	//

	users = userinfo.Scan()                       // scan json tables from boltdb
	usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
	usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
	cleanUserDatabaseWithoutMercy()               // Delete JSON for users that have no passwd entry.
	cleanObjects()                                // Delete orphan files (Objects with no user attached, or objects with no base64 file anymore)
	go srvWebsocketEchoLaunch()
}
func Serve() {
	// Experiment
	listener = waitForListener(*bind, *port)                 // Wait for port to open up
	log.Printf("Got listener: http://%s:%s\n", *bind, *port) // Alert user
	var c = make(chan int)                                   // Auto reboot 1
	go func() {
		if listener != nil {
			http.Serve(listener,
				csrf.Protect(antiCSRFkey,
					csrf.HttpOnly(true),
					csrf.FieldName(sessionName+"-form"),
					csrf.CookieName(sessionName+"-csrf"),
					csrf.Secure(false))(route))
			c <- 1
		}
	}()

	go func() {
		for { //ever
			time.Sleep(10 * time.Second)
			users = userinfo.Scan()                       // scan json tables from boltdb
			usersObjects = userinfo.ScanObjects()         // scan json tables from boltdb
			usersObjectBoxes = userinfo.ScanObjectBoxes() // scan json tables from boltdb
			log.Println("Reloaded [user] [objectbox] and [objects]")
			//	log.Println(users, usersObjectBoxes, usersObjects)
		}
	}()
	select {
	case <-c:
		route = nil
		log.Println("Trying to reboot autopilot!")
		Serve()
		return
	}
}

// waitForListener grabs the wanted port when it becomes available
func waitForListener(bind string, port string) net.Listener {
	originalListener, err1 := net.Listen("tcp", bind+":"+port)
	if err1 != nil {
		panic(err1)
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		panic(err)
	}
	return sl.TCPListener
}
