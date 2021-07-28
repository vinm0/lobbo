package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const (
	PORT = ":8080"

	// templates
	TEMPL_DIR = "templates/"
	HOME      = TEMPL_DIR + "index.html"
	LOBBY     = TEMPL_DIR + "lobby.html"
	PROFILE   = TEMPL_DIR + "profile.html"
	LOBBIES   = TEMPL_DIR + "lobbies.html"
	GROUPS    = TEMPL_DIR + "groups.html"
	NEW_LOBBY = TEMPL_DIR + "lobbyform.html"
	BASE      = TEMPL_DIR + "base.html"

	SITE_TITLE = "Lobbo"
)

type Page struct {
	Title string
}

func launch() {
	fmt.Println("Accessing Homepage")
	http.HandleFunc("/", homeHandler)

	fmt.Println("Launching Server ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadPage(title string) *Page {
	fmt.Println("Homepage Loaded")
	return &Page{Title: title}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Loading Homepage")
	p := loadPage(SITE_TITLE)

	fmt.Println("Parsing Template")
	t, _ := template.ParseFiles(HOME)

	fmt.Println("Executing Template")
	t.Execute(w, p)
}
