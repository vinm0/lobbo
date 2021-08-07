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
	"time"

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
	GROUP_FORM_TEMPL = TEMPL_DIR + "groupform.html"
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
	PRIVATE            = 4

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
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/join/", joinHandler)
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/delete/", deleteHandler)

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

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	path := strings.TrimPrefix(r.URL.Path, "/delete")

	if r.Method != http.MethodPost {
		http.Redirect(w, r, path, http.StatusFound)
	}
	if auth, _ := session[AUTH].(bool); !auth {
		http.Redirect(w, r, "/signin"+path, http.StatusFound)
		return
	}

	id := r.PostFormValue("del-id")
	lby := lobby(id)
	ldr := sessionLeader(session)

	if lby.OwnerID == ldr.LeaderID {
		lby.Delete()
	}

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	path := strings.TrimPrefix(r.URL.Path, "/add")

	if r.Method != http.MethodPost {
		http.Redirect(w, r, path, http.StatusFound)
	}
	if auth, _ := session[AUTH].(bool); !auth {
		http.Redirect(w, r, "/signin"+path, http.StatusFound)
		return
	}

	// ["", "profile", "id"]
	pathSlice := strings.Split(path, "/")

	id := pathSlice[2]
	ldr := sessionLeader(session)

	ldr.AddColleague(id)

	http.Redirect(w, r, path, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	cookie, session := session(r)

	path := r.URL.Path
	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin"+path, http.StatusFound)
		return
	}

	// ["", "edit", "lobby|group", "id"]
	pathSlice := strings.Split(path, "/")

	if len(pathSlice) < 4 {
		fmt.Fprintf(w, "Url not found: %s", path)
		return
	}

	category := pathSlice[2]
	id := pathSlice[3]
	ldr := sessionLeader(session)

	if r.Method == http.MethodPost {
		fmt.Println("Post Method Sent: ", "-"+category+"-")
		fmt.Print("-" + r.PostFormValue("del-id") + "-")
		// field owner_id must be session leader id

		switch category {
		case "lobby":
			owner := ldr.OwnsLobby(id)

			fmt.Println("upd_lobby", "-"+r.PostFormValue("upd_lobby")+"-")

			if r.PostFormValue("upd_lobby") != "" && owner {
				fmt.Println("privacy:", r.PostFormValue("visibility"))
				updateLobby(r.PostForm, id)
			}

			delID := r.PostFormValue("del-id")
			fmt.Println("del-id", "-"+delID+"-")
			if owner || delID == strconv.Itoa(ldr.LeaderID) {
				deleteLobbyMember(id, delID)
			}

			http.Redirect(w, r, strings.TrimPrefix(path, "/edit"), http.StatusFound)
			return

		case "groups":
			updateGroup(r.PostForm, id)
			http.Redirect(w, r, "/groups", http.StatusFound)
			return
		}

	}

	p := &Page{
		"title":  "Edit " + strings.Title(category),
		"leader": ldr,
	}

	var tmpl string

	switch category {
	case "lobby":
		(*p)[category] = lobby(id)
		(*p)["now"] = time.Now()
		tmpl = LOBBY_FORM_TEMPL

	case "groups":
		(*p)["colleagues"] = ldr.ColleaguesAll()
		(*p)["groupname"] = session["groupname"]
		session["groupname"] = nil
		cookie.Save(r, w)
		tmpl = GROUP_FORM_TEMPL
	}

	servePage(w, p, BASE_TEMPL, tmpl)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	_, session := session(r)

	path := r.URL.Path
	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin"+path, http.StatusFound)
		return
	}

	category := strings.TrimPrefix(path, "/new/")
	ldr := sessionLeader(session)

	if r.Method == http.MethodPost {
		// field owner_id must be session leader id
		if r.PostFormValue("owner_id") != strconv.Itoa(ldr.LeaderID) {
			http.Redirect(w, r, path, http.StatusFound)
		}

		var newID int
		switch category {
		case "lobby":
			newID = updateLobby(r.PostForm, "")

			newPath := fmt.Sprintf("/%s/%d", category, newID)
			http.Redirect(w, r, newPath, http.StatusFound)
			return
		case "groups":
			fmt.Println("Post form\n", r.PostForm)
			updateGroup(r.PostForm, "")
			http.Redirect(w, r, "/groups", http.StatusFound)
			return
		}

	}

	p := &Page{
		"title":      "New " + strings.Title(category),
		"leader":     ldr,
		"colleagues": ldr.ColleaguesAll(),
	}

	var tmpl string

	switch category {
	case "lobby":
		(*p)[category] = &Lobby{}
		tmpl = LOBBY_FORM_TEMPL

	case "groups":
		(*p)[category] = &Group{}
		tmpl = GROUP_FORM_TEMPL
	}

	servePage(w, p, BASE_TEMPL, tmpl)
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, session := session(r)

	if auth, _ := session["authenticated"].(bool); !auth {
		http.Redirect(w, r, "/signin/groups", http.StatusFound)
		return
	}

	ldr := sessionLeader(session)

	if r.Method == http.MethodPost {
		delID := r.PostFormValue("del_id")
		if delID != "" {
			id, _ := strconv.Atoi(delID)
			gid := r.PostFormValue("grp_id")
			groupID, _ := strconv.Atoi(gid)

			if !ldr.ownsGroup(groupID) {
				fmt.Fprintln(w, "Unauthorized to edit group: ", groupID)
			}

			deleteGroupMember(groupID, id)

			http.Redirect(w, r, "/groups/#grp"+gid, http.StatusFound)
			return
		}

		gpID := r.PostFormValue("group_id")
		if gpID != "" {
			session["groupname"] = r.PostFormValue("groupname")
			cookie.Save(r, w)
			http.Redirect(w, r, "/edit/groups/"+gpID, http.StatusFound)
			return
		}
	}

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
	ldr := sessionLeader(session)

	var owner *Leader
	isOwner := lby.OwnerID == ldr.LeaderID
	if isOwner {
		owner = ldr
	} else {
		owner = lby.Owner()
	}

	// TODO: Fix valid privacy verification
	if lby.OwnerID != session["leader_id"] &&
		lby.Privacy > FRIENDS_OF_FRIENDS {
		fmt.Fprint(w, "Access Denied")
	}

	p := &Page{
		"title":   lby.Title,
		"lobby":   lby,
		"members": members(lby.LobbyID),
		"owner":   owner,
		"leader":  ldr,
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

	prof := LeaderProfile(session, path)
	ldr := sessionLeader(session)
	isOwner := prof.LeaderID == ldr.LeaderID

	// session.Values[]
	p := &Page{
		"title":        PROFILE_TITLE,
		"leader":       prof,
		"leader_id":    ldr.LeaderID,
		"ownedLobbies": ownedLobbies(prof.LeaderID, 10),
		"inLobbies":    inLobbies(prof.LeaderID, 10),
		"colleagues":   prof.Colleagues(10),
		"is_colleague": ldr.IsColleague(prof.LeaderID) || isOwner,
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

func updateLobby(form url.Values, lobbyID string) (newID int) {
	form["meet_time"][0] = form.Get("meet_date") + " " + form.Get("meet_time")

	if lobbyID == "" {
		return CreateLobbyDB(form)
	}

	id, _ := strconv.Atoi(lobbyID)
	UpdateLobbyDB(form, id)

	return 0
}

func updateGroup(form url.Values, groupID string) (newID int) {
	if groupID == "" {
		newID = CreateGroupDB(form)
		if form.Get("colleagues") != "" {
			AddGroupMembersDB(form, newID)
		}
		return newID
	}

	id, _ := strconv.Atoi(groupID)
	UpdateGroupDB(form, id)
	if form.Get("colleagues") != "" {
		AddGroupMembersDB(form, id)
	}

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

func deleteGroupMember(groupID int, memberID int) {
	DeleteGroupMemberDB(groupID, memberID)
}

func deleteLobbyMember(lobbyID string, memberID string) {
	lID, _ := strconv.Atoi(lobbyID)
	mID, _ := strconv.Atoi(memberID)
	DeleteLobbyMemberDB(lID, mID)
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
