package main

import (
	"ircclient"
	"log"
//	"time"
)

func main() {
	//for _, line := range server_lines {
	//	ircclient.ParseServerLine(line)
	//}
	s := ircclient.NewIRCClient("dpaulus.dyndns.org:6667", "testbot", "testbot", "ident", '.')
	ok := s.Connect()
	if ok != nil {
		log.Fatal(ok.String())
	}
	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.String())
	}
	//s.Output <- "Hello, world\n"
	//s.Output <- "Asdf!\n"
	//s.Quit()
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
