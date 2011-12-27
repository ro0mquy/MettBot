package main

import (
	"ircconn"
	"fmt"
)

func main() {
	s := ircconn.NewIRCConn()
	s.Connect("localhost:6667")

	serverData := s.RegisterListener()
	go output_channel(serverData)

	clientInput := s.RegisterWriter()
	st := ""
	for {
		fmt.Scanln(st)
		clientInput <- st
	}
}

func output_channel(c <-chan string) {
	for {
		s := <- c
		fmt.Println("-> " + s)
	}
}
