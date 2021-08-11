package main

import "log"

// Accepts an error variable and checks whether an error exists.
// If an error exists, the error is printed to the log.
// Items provide context for where the error may have occurred.
func Check(err error, items ...interface{}) {
	if err != nil {
		log.Println(err.Error())
		log.Print(items...)
	}
}
