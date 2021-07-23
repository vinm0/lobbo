package main

import (
	"os"
	"time"
)

const (
	tmplDir = "templates/"
	baseTempl = tmplDir + "base.html"
	profileTempl = tmplDir + "profile.html"
	groupsTempl = tmplDir + "groups.html"
	homeTempl = tmplDir + "index.html"
	lobbyTempl = tmplDir + "lobby.html"
	lobbiesTempl = tmplDir + "lobbies.html"
	lobbyformTempl = tmplDir + "lobbyform.html"
)

const (
	arrSize = 5
	stayLoggedInBlock = 14
	isLoggedInBlock = 12
	lastLoginBlock = 11
	usrnameBlock = 10
	pwordBlock = 10
)

const(
	inviteOnlyMin = 2
	maxLobbyFree = 3
	maxLobbyPaid = 10
	maxLobbyPremium = 20
	maxLobbySearch = 10
)


var(
	usrname []byte
	pword []byte
	
	home, _ = os.UserHomeDir()
	localDir = home + "/lobbo/"
	loginFileName = "login.txt"
	loginFilePath = localDir + loginFileName


	loginBool = []byte(
		`stayLoggedIn: true
		isLoggedIn: true`)

	buildLogin = [][]byte {
		[]byte("stayLoggedIn: "), 
		[]byte("isLoggedIn: "), 
		[]byte("lastLogin: "), 
		[]byte("username: "), 
		[]byte("password: ")}
)

type person struct {
	personID     int
	username	 string
	firstName    string
	lastName	 string
	email		 string
	profile	 	 string
	bio			 string
	socials		 map[string]string  // [platform]link
	lobbiesOwned []lobby
	lobbiesIn    []lobby
	groups		 []group
	acctType	 int
}

type client struct {
	person
	isClient	bool
}

type lobby struct {
	lobbyID       int
	owner         person
	name          string
	description	  string
	activityDate  time.Time
	location	  string
	repeat		  string
	members		  []person
	capacity      int
	memberCount   int
	isInviteOnly  bool
	privacy		  int 
	dateCreated   time.Time
}

type group struct {
	groupID		  int
	name		  string
	owner		  int
	members		  []person
}

type file struct {
	path string
	name string
	contents []byte
}