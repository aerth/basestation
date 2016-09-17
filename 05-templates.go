package basestation

import (
	"html/template"
	"log"
	"net/http"
)

var viewInfo sessionView

func newView(req *http.Request) *sessionView {
	v := &sessionView{}
	v.Vars = make(map[string]interface{})
	v.BaseURI = viewInfo.BaseURI
	v.Extension = viewInfo.Extension
	v.Folder = viewInfo.Folder
	v.Name = viewInfo.Name
	// Make sure BaseURI is available in the templates
	v.Vars["BaseURI"] = v.BaseURI
	// This is required for the view to access the request
	v.request = req
	// Get session
	//sess := session.Instance(v.request)
	return v
}
func id2email(id string) string {
	user, err := getUserDataByID(id)
	if err != nil {
		return "[ghost]"
	}
	return user.Email
}
func id2name(id string) string {
	user, errar := getUserDataByID(id)
	if errar != nil {
		log.Println(errar)
		return "[ghost]"
	}
	return user.NickName
}
func email2id(email string) string {
	_, id, errar := getUserDataByEmail(email)
	if errar != nil {
		log.Println(errar)
		return "0"
	}
	return id
}
func name2id(name string) string {
	for _, user := range users {
		if user.NickName == name {
			return user.ID
		}
	}
	return "0"
}

func loadTemplate(name string) (*template.Template, error) {
	filename := "/" + name + ".html"
	oth := templateDir + "/header.html"
	ers := templateDir + "/footer.html"
	log := templateDir + "/login.html"
	sin := templateDir + "/signup.html"

	funcMap := template.FuncMap{
		"id2name":  id2name,
		"id2email": id2email,
		"email2id": email2id,
		"name2id":  name2id,
	}
	t, err := template.New(name).Funcs(funcMap).ParseFiles(templateDir+filename, oth, ers, log, sin)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func triplecombo(w http.ResponseWriter, s string, data map[string]interface{}) {
	t, e := loadTemplate(s)
	if e != nil {
		panic(e)
	}
	t.ExecuteTemplate(w, "header", data)
	t.ExecuteTemplate(w, s, data)
	t.ExecuteTemplate(w, "footer", data)

}
func runTemplate(s string) {

}
