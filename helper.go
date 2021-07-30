package main

import "log"

func Check(err error, items ...interface{}) {
	if err != nil {
		log.Println(err.Error())
		log.Print(items...)
	}
}
