package main

import (
	"fmt"
	"log"
	"os"
)

var accessToken = os.Getenv("ACCESS_TOKEN")

func main() {
	c := NewClient(accessToken)
	// FIXME named args
	fileMembers, err := c.ListFileMembers("id:MP8Ja5KjILAAAAAAAAAACg", true, 1)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(fileMembers)
}
