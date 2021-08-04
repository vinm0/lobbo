package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	TEMPL_DIR        = "templates/"
	HOME_TEMPL       = TEMPL_DIR + "index.html"
	SIGNIN_TEMPL     = TEMPL_DIR + "signin.html"
	LOBBY_TEMPL      = TEMPL_DIR + "lobby.html"
	PROFILE_TEMPL    = TEMPL_DIR + "profile.html"
	LOBBIES_TEMPL    = TEMPL_DIR + "lobbies.html"
	GROUPS_TEMPL     = TEMPL_DIR + "groups.html"
	LOBBY_FORM_TEMPL = TEMPL_DIR + "lobbyform.html"
	BASE_TEMPL       = TEMPL_DIR + "base.html"

	SITE_TITLE      = "Lobbo"
	SIGNIN_TITLE    = "Sign-in"
	PROFILE_TITLE   = "Profile"
	LOBBIES_TITLE   = "Lobbies"
	NEW_LOBBY_TITLE = "New"
	GROUPS_TITLE    = "Groups"
	INBOX_TITLE     = "Inbox"

	// Cookie Names
	SESSION = "session"
	AUTH    = "authenticated"
	PASS    = "pass"
	LDR_ID  = "leader_id"
	LNAME   = "lname"
	FNAME   = "fname"
	EMAIL   = "email"

	// Privacy Levels
	PUBLIC             = 0
	COMMUNITY          = 1
	FRIENDS_OF_FRIENDS = 2
	FRIENDS            = 3
	PRIVATE

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
	http.HandleFunc("/lobby/", lobbyHandler)
	http.HandleFunc("/lobbies/", lobbiesHandler)
	http.HandleFunc("/lobbies-in/", lobbiesHandler)
	http.HandleFunc("/groups/", groupsHandler)
	http.HandleFunc("/edit/", lobbyFormHandler)
	http.HandleFunc("/new/", lobbyFormHandler)
	http.HandleFunc("/join/", joinHandler)
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

func lobbyFormHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	path := r.URL.Path

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin"+path, http.StatusFound)
		return
	}

	new := path == "/new/"

	var id string
	if !new {
		//["", "new|edit", "lobby", "id"]
		id = strings.Split(path, "/")[3]
	}

	ldr := sessionLeader(session)

	if r.Method == http.MethodPost {
		if !new && r.PostFormValue("owner_id") != strconv.Itoa(ldr.LeaderID) {
			fmt.Fprintln(w, "Not authorized to edit lobby: ", id)
			return
		}

		r.ParseForm()
		r.PostForm["meet_time"][0] = r.PostForm["meet_date"][0] + " " + r.PostForm["meet_time"][0]
		newID := updateLobby(r.PostForm, id, new)

		if newID != 0 {
			id = strconv.Itoa(newID)
		}

		http.Redirect(w, r, "/lobby/"+id, http.StatusFound)
		return
	}

	if new {
		p := &Page{
			"title":     "New Lobby",
			"lobby":     &Lobby{},
			"leader_id": ldr.LeaderID,
		}
		fmt.Println("Session LeaderID: ", session[LDR_ID])
		servePage(w, p, BASE_TEMPL, LOBBY_FORM_TEMPL)
		return
	}

	lby := lobby(id)

	if lby.LobbyID == 0 || !ldr.isOwner(lby) {
		fmt.Fprintln(w, "Unable to edit lobby: ", id)
		return
	}

	p := &Page{
		"title":     lby.Title,
		"lobby":     lby,
		"leader_id": ldr.LeaderID,
	}

	servePage(w, p, BASE_TEMPL, LOBBY_FORM_TEMPL)
}

// TODO: load groups data. Update database
func groupsHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin/groups", http.StatusFound)
		return
	}

	ldr := sessionLeader(session)
	grp := ldr.Groups()

	p := &Page{
		"title":  "Groups",
		"leader": ldr,
		"groups": grp,
	}

	fmt.Println(grp)
	servePage(w, p, BASE_TEMPL, GROUPS_TEMPL)
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	id := strings.TrimPrefix(r.URL.Path, "/join/")

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin/join/"+id, http.StatusFound)
		return
	}

	leaderID, _ := session[LDR_ID].(int)
	lobbyID, _ := strconv.Atoi(id)
	prv, _ := strconv.Atoi(r.PostFormValue("join-invite"))

	if !joinAllowed(lobbyID, leaderID, prv) {
		fmt.Fprint(w, "Permission to join lobby denied")
		return
	}

	joinLobby(lobbyID, leaderID)

	http.Redirect(w, r, "/lobby/"+id, http.StatusFound)
}

func lobbyHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	lobbyID := strings.TrimPrefix(r.URL.Path, "/lobby/")
	lby := lobby(lobbyID)
	owner := lobbyOwner(session, lby.OwnerID)

	// TODO: Fix valid privacy verification
	if lby.OwnerID != session["leader_id"] &&
		lby.Privacy > FRIENDS_OF_FRIENDS {
		fmt.Fprint(w, "Access Denied")
	}

	p := &Page{
		"title":     lby.Title,
		"lobby":     lby,
		"members":   members(lby.LobbyID),
		"owner":     owner,
		"leader_id": session[LDR_ID],
	}

	// TODO: 404 if loby doesn't exist
	servePage(w, p, BASE_TEMPL, LOBBY_TEMPL)
}

func lobbiesHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		if lobbyType := r.PostFormValue("lobby-type"); lobbyType != "" {
			switch lobbyType {
			case "My Lobbies":
				http.Redirect(w, r, "/lobbies", http.StatusFound)

			case "Lobbies In":
				http.Redirect(w, r, "/lobbies-in", http.StatusFound)
			}
			return
		}

		if lobbyID := r.PostFormValue("lobby-id"); lobbyID != "" {
			http.Redirect(w, r, "/lobbyform/"+lobbyID, http.StatusFound)
			return
		}
	}

	ldr := sessionLeader(session)
	lbys := loadLobbies(r, ldr.LeaderID)

	p := &Page{
		"title":   strings.Title(strings.Trim(r.URL.Path, "/")),
		"lobbies": lbys,
	}

	servePage(w, p, BASE_TEMPL, LOBBIES_TEMPL)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

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

	path := strings.TrimPrefix(r.URL.Path, "/profile/")

	ldr := LeaderProfile(session, path)

	// session.Values[]
	p := &Page{
		"title":        PROFILE_TITLE,
		"leader":       ldr,
		"leader_id":    session[LDR_ID],
		"ownedLobbies": ownedLobbies(ldr.LeaderID, 10),
		"inLobbies":    inLobbies(ldr.LeaderID, 10),
		"colleagues":   colleagues(ldr.LeaderID, 10),
	}

	servePage(w, p, BASE_TEMPL, PROFILE_TEMPL)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	cookie, session := session(r)

	if r.Method == http.MethodGet {
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

	// redirect to previous page
	if len(r.URL.Path) > len("/signin/") {
		path := strings.TrimPrefix(r.URL.Path, "/signin")
		fmt.Println("redirecting to ", path)
		http.Redirect(w, r, path, http.StatusFound)
		return
	}
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
	// fmt.Println("Parsing templates")
	t := template.Must(template.ParseFiles(templ...))

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

func sessionLeader(session map[interface{}]interface{}) *Leader {
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

func updateLobby(form url.Values, lobbyID string, new bool) (newID int) {
	if new {
		newID = CreateLobbyDB(form)
		return newID
	}
	id, _ := strconv.Atoi(lobbyID)
	UpdateLobbyDB(form, id)
	return 0
}

func joinLobby(lobbyID int, leaderID int) {
	JoinLobbyDB(lobbyID, leaderID)
}

func LeaderProfile(session map[interface{}]interface{}, path string) *Leader {
	id, _ := strconv.Atoi(path)

	if path == "" || id == session["leader-id"] {
		return sessionLeader(session)
	}

	return LeaderDB(id)
}

func loadLobbies(r *http.Request, leaderID int) []*Lobby {
	if r.URL.Path == "/lobbies-in/" {
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

func lobby(lobbyID string) *Lobby {
	id, _ := strconv.Atoi(lobbyID)
	return LobbyDB(id)
}

func members(lobbyID int) []*Leader {
	return MembersDB(lobbyID)
}

func lobbyOwner(session map[interface{}]interface{}, ownerID int) *Leader {
	if session[LDR_ID] == ownerID {
		return sessionLeader(session)
	}

	return LeaderDB(ownerID)
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

func joinAllowed(lobbyID int, leaderID int, inviteCode int) bool {
	return inviteCode <= 2
	// TODO verify permissions based on network
	// return JoinAllowedDB(lobbyID, leaderID)
}

func isEmail(email string) bool {
	valid, err := regexp.MatchString(`.{3,}@.+\..+$`, email)
	Check(err, "Email match error:", email)

	return valid
}

func cleanString(s string) string {
	return strings.ToLower(strings.Trim(s, " "))
}
