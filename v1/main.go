package main

import (
	"log"
	"net/http"
)



func main() {
    var c client

    http.HandleFunc("/", c.homeHandler)
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}