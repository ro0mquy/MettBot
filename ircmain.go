package main

import (
	"ircconn"
	//	"fmt"
)

func main() {
	s := ircconn.NewIRCConn()
	s.Connect("localhost:6667")
	s.Output <- "Hello, world\n"
	s.Output <- "Asdf!\n"
	s.Quit()
	//	go output_channel(serverData)
	//	for {
	//		fmt.Scanln(st)
	//		clientInput <- st
	//	}
}

func output_channel(c <-chan string) {
	//	for {
	//		s := <- c
	//		fmt.Println("-> " + s)
	//	}
}
