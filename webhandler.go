package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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

	MIN_PWD_LEN = 8
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
	http.HandleFunc("/signout/", signoutHandler)
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
	for k, v := range session.Values {
		fmt.Println("key:", k, "\nvalue: ", v)
	}
	// 	"id: %d\nusr: %s\nfname: %s\nlname: %s",
	// 	session.Values[LDR_ID],
	// 	session.Values[EMAIL],
	// 	session.Values[FNAME],
	// 	session.Values[LNAME],
	// )
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)

	if r.Method == "GET" {
		fmt.Println("Get request")
		// if auth, _ := session.Values["authenticated"].(bool); auth {
		// 	http.Redirect(w, r, "/profile", http.StatusFound)
		// 	return
		// }

		servePage(w, SIGNIN_TITLE, SIGNIN_TEMPL)
		return
	}

	usr := cleanString(r.PostFormValue(EMAIL))
	pwd := cleanString(r.PostFormValue(PASS))

	if valid, msg := validateSignin(usr, pwd); !valid {
		session.Values["error"] = msg
		session.Save(r, w)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

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

func signoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)
	for k := range session.Values {
		delete(session.Values, k)
	}
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func validateSignin(usr string, pwd string) (valid bool, msg string) {
	s := []string{}
	if len(usr) < 1 && !isEmail(usr) {
		s = append(s, "Please enter a valid email.")
	}
	if len(pwd) < MIN_PWD_LEN {
		s = append(s, "Password must be 8 or more characters")
	}

	return len(s) == 0, strings.Join(s, "<br>")
}

func isEmail(email string) bool {
	valid, err := regexp.MatchString(`.{3,}@.+\..+$`, email)
	Check(err, "Email match error:", email)

	return valid
}

func cleanString(s string) string {
	return strings.ToLower(strings.Trim(s, " "))
}
