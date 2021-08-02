package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

	SITE_TITLE      = "Lobbo"
	SIGNIN_TITLE    = "Sign-in"
	PROFILE_TITLE   = "Profile"
	LOBBIES_TITLE   = "Lobbies"
	NEW_LOBBY_TITLE = "New"
	GROUPS_TITLE    = "Groups"
	INBOX_TITLE     = "Inbox"

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

type Page map[string]interface{}

var store *sessions.CookieStore

func init() {
	godotenv.Load(".env")
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
}

func launch() {

	http.HandleFunc("/signin/", signinHandler)
	http.HandleFunc("/signout/", signoutHandler)
	http.HandleFunc("/profile/", profileHandler)
	http.HandleFunc("/lobbies/", lobbiesHandler)
	http.HandleFunc("/lobbies-in/", lobbiesHandler)
	http.HandleFunc("/", homeHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	fmt.Println("Launching Server on port", PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))
}

/*
 * ********************************************************************
 * ********************************************************************
 * *********************** Handler Functions **************************
 * ********************************************************************
 * ********************************************************************
 */

func homeHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{"title": SITE_TITLE}
	servePage(w, p, HOME_TEMPL)
}

func lobbiesHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		switch r.PostFormValue("lobby-type") {
		case "My Lobbies":
			http.Redirect(w, r, "/lobbies", http.StatusFound)

		case "Lobbies In":
			http.Redirect(w, r, "/lobbies-in", http.StatusFound)
		}
		return
	}

	ldr := loadLeader(session)
	lbys := loadLobbies(r, ldr.LeaderID)

	p := &Page{
		"title":   strings.Title(strings.Trim(r.URL.Path, "/")),
		"lobbies": lbys,
	}

	servePage(w, p, BASE_TEMPL, LOBBIES_TEMPL)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	fmt.Println("session:", session)

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost && r.FormValue("delete") == "Delete" {
		ownerID, _ := session["leader_id"].(int)
		colleagueID, _ := strconv.Atoi(r.FormValue("del-id"))

		deleteColleague(ownerID, colleagueID)
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	fmt.Println("Loading user info")
	ldr := loadLeader(session)

	// session.Values[]
	fmt.Println("Loading Page context")
	p := &Page{
		"title":        PROFILE_TITLE,
		"leader":       ldr,
		"ownedLobbies": ownedLobbies(ldr.LeaderID, 10),
		"inLobbies":    inLobbies(ldr.LeaderID, 10),
		"colleagues":   colleagues(ldr.LeaderID, 10),
	}

	fmt.Println("serving page")
	servePage(w, p, BASE_TEMPL, PROFILE_TEMPL)
	fmt.Println("Page served:", PROFILE_TEMPL, BASE_TEMPL)

	// for k, v := range session.Values {
	// 	fmt.Println("key:", k, "\nvalue: ", v)
	// }
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	cookie, session := session(r)

	if r.Method == http.MethodGet {
		fmt.Println("Get request")
		if auth, _ := session["authenticated"].(bool); auth {
			http.Redirect(w, r, "/profile", http.StatusFound)
			return
		}

		p := &Page{"title": SIGNIN_TITLE}
		servePage(w, p, SIGNIN_TEMPL)
		return
	}

	usr := cleanString(r.PostFormValue(EMAIL))
	pwd := cleanString(r.PostFormValue(PASS))

	if valid, errMsg := validateSignin(usr, pwd); !valid {
		session["error"] = errMsg
		cookie.Save(r, w)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	ldr, err := Auth(usr, pwd)
	Check(err, "login err for user: ", usr)

	session[FNAME] = ldr.Firstname
	session[LNAME] = ldr.Lastname
	session[EMAIL] = usr
	session[LDR_ID] = ldr.LeaderID
	session[AUTH] = true
	cookie.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func signoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION)
	session.Values = make(map[interface{}]interface{})
	session.Save(r, w)

	http.Redirect(w, r, "/signin", http.StatusFound)
}

/*
 * ********************************************************************
 * ********************************************************************
 * *********************** Helper Functions ***************************
 * ********************************************************************
 * ********************************************************************
 */

func servePage(w http.ResponseWriter, p *Page, templ ...string) {
	fmt.Println("Parsing templates")
	t := template.Must(template.ParseFiles(templ...))
	fmt.Println((*t).Tree)
	// if err != nil {
	// 	log.Println("Unable to parse file:", templ)
	// 	log.Println(err.Error())
	// }
	fmt.Println("Executing templates")
	err := t.Execute(w, p)
	if err != nil {
		log.Println(err.Error(), "Unable to execute template.")
	}
}

func session(r *http.Request) (cookie *sessions.Session, session map[interface{}]interface{}) {
	cookie, _ = store.Get(r, SESSION)
	return cookie, cookie.Values
}

func loadLeader(session map[interface{}]interface{}) *Leader {
	id, _ := session[LDR_ID].(int)
	em, _ := session[EMAIL].(string)
	fn, _ := session[FNAME].(string)
	ln, _ := session[LNAME].(string)
	return &Leader{
		LeaderID:  id,
		Username:  em,
		Firstname: fn,
		Lastname:  ln}
}

func loadLobbies(r *http.Request, leaderID int) []*Lobby {
	if r.URL.Path == "lobbies-in/" {
		return inLobbiesAll(leaderID)
	}

	return ownedLobbiesAll(leaderID)
}

func ownedLobbies(ownerID int, limit int) []*Lobby {
	return OwnedLobbiesDB(ownerID, " Limit "+strconv.Itoa(limit))
}

func ownedLobbiesAll(ownerID int) []*Lobby {
	return OwnedLobbiesDB(ownerID, "")
}

func inLobbies(memberID int, limit int) []*Lobby {
	return inLobbiesDB(memberID, " Limit "+strconv.Itoa(limit))
}

func inLobbiesAll(memberID int) []*Lobby {
	return inLobbiesDB(memberID, "")
}

func colleagues(ownerID int, limit int) []*Leader {
	return ColleaguesDB(ownerID, " Limit "+strconv.Itoa(limit))
}

func deleteColleague(ownerID int, colleagueID int) {
	DeleteColleagueDB(ownerID, colleagueID)
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
