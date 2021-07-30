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
	TEMPL_DIR       = "templates/"
	HOME_TEMPL      = TEMPL_DIR + "index.html"
	SIGNIN_TEMPL    = TEMPL_DIR + "signin.html"
	LOBBY_TEMPL     = TEMPL_DIR + "lobby.html"
	PROFILE_TEMPL   = TEMPL_DIR + "profile.html"
	LOBBIES_TEMPL   = TEMPL_DIR + "lobbies.html"
	GROUPS_TEMPL    = TEMPL_DIR + "groups.html"
	NEW_LOBBY_TEMPL = TEMPL_DIR + "lobbyform.html"
	BASE_TEMPL      = TEMPL_DIR + "base.html"

	SITE_TITLE   = "Lobbo"
	SIGNIN_TITLE = "Sign-in"

	// cookies
	SESSION = "session"
	AUTH    = "authenticated"
	PASS    = "pass"
	LDR_ID  = "leader_id"
	LNAME   = "lname"
	FNAME   = "fname"
	EMAIL   = "email"
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

func servePage(w http.ResponseWriter, title string, templ string) {
	p := loadPage(title)
	t, err := template.ParseFiles(templ)
	if err != nil {
		log.Printf("Unable to parse file: %s. \n", templ)
		log.Println(err.Error())
	}
	t.Execute(w, p)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	p := loadPage(SITE_TITLE)
	t, _ := template.ParseFiles(HOME_TEMPL)

	t.Execute(w, p)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)
	fmt.Printf("id: %d\nusr: %s\nfname: %s\nlname: %s",
		session.Values[LDR_ID],
		session.Values[EMAIL],
		session.Values[FNAME],
		session.Values[LNAME],
	)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)

	if r.Method == "GET" {
		fmt.Println("Get request")
		if auth, _ := session.Values["authenticated"].(bool); auth {
			http.Redirect(w, r, "/profile", http.StatusFound)
			return
		}

		servePage(w, SIGNIN_TITLE, SIGNIN_TEMPL)
		return
	}

	fmt.Println("Post request")
	usr := r.PostFormValue(EMAIL)
	pwd := r.PostFormValue(PASS)
	fmt.Printf("usr: %s\npwd: %s", usr, pwd)

	ldr, err := Auth(usr, pwd)
	Check(err, "login err for user: ", usr)

	session.Values[FNAME] = ldr.Firstname
	session.Values[LNAME] = ldr.Lastname
	session.Values[EMAIL] = usr
	session.Values[LDR_ID] = ldr.LeaderID
	session.Values[AUTH] = true
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusFound)
}
