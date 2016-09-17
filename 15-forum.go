package basestation

import (
	"log"
	"net/http"
	"strings"

	"github.com/aerth/fforum"
	"github.com/gorilla/csrf"
)

func init() {

}
func srvForum(w http.ResponseWriter, r *http.Request) {

	sess, user := roadblock(w, r)
	if user.ID == "" {
		return
	}
	v := newView(r)
	if flashes := sess.Flashes(); len(flashes) > 0 {
		v.Vars["flashes"] = flashes
	}
	sess.Save(r, w)
	if sess.Values["ACTIVE"] != true {
		redirect(w, r, "dashboard")
	}

	if r.Method == "POST" {

		if !validateForum(r) {
			sess.AddFlash("Invalid topic or reply. Not published.")
			sess.Save(r, w)
			redirect(w, r, "/forum")
			return
		}

		switch r.FormValue("action") {
		case "newtopic":
			if r.FormValue("body") == "" {
				redirect(w, r, "/forum")
				return
			}
			log.Println("got newtopic")
			p := forum.NewTopic()
			p.Body = r.FormValue("body")
			p.Owner = user.ID
			p.Title = r.FormValue("title")
			p.Category = r.FormValue("category")

			log.Printf("saving id %s, category %s, owner %s, body %q", p.ID, p.Category, p.Owner, p.Body)
			err = p.Save()
			if err != nil {
				log.Println(err)
			}
			log.Println("saved it?")
		case "newreply":
			if r.FormValue("body") == "" {
				redirect(w, r, "/forum")
				return
			}
			if r.FormValue("to") == "" {
				redirect(w, r, "/forum")
				return
			}
			log.Println("got newreply")
			t, err := forum.ReadTopic(r.FormValue("to"))
			if err != nil {
				log.Println(err)
				redirect(w, r, "/forum")
				return
			}
			reply := t.NewReply()
			reply.Body = r.FormValue("body")
			reply.Owner = user.ID
			log.Printf("saving reply id %s,  to %s,  owner %s, body %q", reply.ID, t.ID, reply.Owner, reply.Body)
			err = reply.Save()
			if err != nil {
				log.Println(err)
			}
			log.Println("saved it?")
			redirect(w, r, "/forum/t/"+t.ID)
			return
		case "newcategory":
			log.Println("got new category")
			c := forum.NewCategory()
			c.Name = r.FormValue("category")
			c.Creator = user.ID
			err = c.Save()
			if err != nil {
				log.Println(err)
			}

			//   redirect(w, r, "/forum")
			//   return
			// case "deletepost":
			//   forum.deletepost(user.ID, r.FormValue("id"))
			//   redirect(w, r, "/forum")
			//   return
			// case "deletecategory":
			//   forum.deletecategory(user.ID, r.FormValue("categoryname"))
			//   redirect(w, r, "/forum")
			//   return

		}
		redirect(w, r, "/forum")
		return
	}

	// Here we build a forum using map[string]OurTypes for the html/template package
	var result *forum.Topic
	var results = map[string]forum.Topic{}
	replies := map[string]forum.Reply{}
	switch { // Switch hits the first match
	case strings.Contains(r.URL.Path, "/forum/t/"): // Is a topic
		topicid := r.URL.Path[len("/forum/t/"):]
		topic, err := forum.ReadTopic(topicid)
		if err != nil {
			log.Println(err)
		}
		result = topic
	case strings.Contains(r.URL.Path, "/forum/r/"): // Is a reply
		rid := r.URL.Path[len("/forum/r/"):]
		reply := forum.ReadReply(rid)
		if reply.ID != "" {
			replies[rid] = reply
		}
	case strings.Contains(r.URL.Path, "/forum/c/"): // Is a category
		catid := r.URL.Path[len("/forum/c/"):]
		results = forum.ListTopicsOf(catid)
	}

	if result != nil {
		replies = result.Replies()
	}
	tops := forum.ListTopicsAll()
	cats := forum.ListCategories()
	data := map[string]interface{}{
		"MyFlashes":      v.Vars["flashes"],
		"Vars":           v.Vars,
		"MyUserData":     user,
		"Categories":     cats,
		"Topics":         tops,
		"Result":         result,
		"Results":        results,
		"Replies":        replies,
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	triplecombo(w, "forum", data)
}
func validateForum(r *http.Request) bool {
	e := r.ParseForm()
	if e != nil {
		log.Println(e)
		return false
	}
	if r.FormValue("action") == "" {
		return false
	}
	switch r.FormValue("action") {
	case "newtopic":
		if r.FormValue("title") == "" || r.FormValue("body") == "" || r.FormValue("category") == "" {
			return false
		}
	case "newreply":
		if r.FormValue("to") == "" || r.FormValue("body") == "" {
			return false
		}
	case "newcategory":
		if r.FormValue("category") == "" {
			log.Println("blank category")
			return false
		}
	default:
		return false
	}
	return true

}
