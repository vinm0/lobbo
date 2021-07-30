package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

const (
	PORT = ":8080"

	// templates
	TEMPL_DIR = "templates/"
	HOME      = TEMPL_DIR + "index.html"
	SIGN_IN   = TEMPL_DIR + "signin.html"
	LOBBY     = TEMPL_DIR + "lobby.html"
	PROFILE   = TEMPL_DIR + "profile.html"
	LOBBIES   = TEMPL_DIR + "lobbies.html"
	GROUPS    = TEMPL_DIR + "groups.html"
	NEW_LOBBY = TEMPL_DIR + "lobbyform.html"
	BASE      = TEMPL_DIR + "base.html"

	SITE_TITLE = "Lobbo"

	// cookies
	SESSION = "session"
)

type Page struct {
	Title string
}

var store *sessions.CookieStore

func init() {
	godotenv.Load(".env")
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}

func launch() {

	http.HandleFunc("/signin/", signinHandler)
	http.HandleFunc("/profile/", profileHandler)
	http.HandleFunc("/", homeHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Launching Server on port", PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))
}

func loadPage(title string) *Page {
	return &Page{Title: title}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	p := loadPage(SITE_TITLE)
	t, _ := template.ParseFiles(HOME)

	t.Execute(w, p)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)
	fmt.Printf("id: %d\nusr: %s\nfname: %s\nlname: %s",
		session.Values["leader_id"],
		session.Values["email"],
		session.Values["fname"],
		session.Values["lname"],
	)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)

	// auth, _ := session.Values["authenticated"].(bool)

	if r.Method == "GET" {
		fmt.Println("Get request")
		// if auth {
		// 	session.Values["authenticated"] = false
		// 	http.Redirect(w, r, "/profile", http.StatusFound)
		// 	return
		// }

		p := loadPage("login")
		t, err := template.ParseFiles(SIGN_IN)
		if err != nil {
			log.Printf("Unable to parse file: %s. \n", SIGN_IN)
			log.Println(err.Error())
		}
		t.Execute(w, p)
		return
	}

	fmt.Println("Post request")
	usr := r.PostFormValue("email")
	pwd := r.PostFormValue("pass")
	fmt.Printf("usr: %s\npwd: %s", usr, pwd)

	ldr, err := Auth(usr, pwd)
	if err != nil {
		log.Printf("Cannot load user: %s. \n", usr)
		log.Println(err.Error())
	}

	session.Values["fname"] = ldr.Firstname
	session.Values["lname"] = ldr.Lastname
	session.Values["email"] = ldr.Username
	session.Values["leader_id"] = ldr.LeaderID
	session.Values["authenticated"] = true
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}
