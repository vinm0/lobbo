package main

import (
	"html/template"
	"net/http"
)

func (c *client) loadData() {
	
}

func renderTemplate(w http.ResponseWriter, tmpl string, c *client) {
    t, _ := template.ParseFiles("templates" + tmpl + ".html")
    t.Execute(w, c)
}

func (c *client) homeHandler(w http.ResponseWriter, r *http.Request) {
    if !c.isClient {
		template.ParseFiles("templates/index.html")
	} else {
		renderTemplate(w, "profile", c)
	}
}

func (c *client) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !c.isClient {
		template.ParseFiles("templates/login.html")
	} else {
		renderTemplate(w, "profile", c)
	}
}



