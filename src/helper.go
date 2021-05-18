package main

import (
	"os"
	"bufio"
)


func scanner() string {
	x := bufio.NewScanner(os.Stdin)
	x.Scan()
	return  x.Text()
}
