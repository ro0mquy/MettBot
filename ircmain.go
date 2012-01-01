package main

import (
	"ircclient"
	"log"
	"github.com/kless/goconfig/config"
)

func main() {
	c, ok := config.ReadDefault("go-faui2k11.cfg")
	if ok != nil {
		c = config.NewDefault()
		c.AddSection("Server")
		c.AddOption("Server", "host", "dpaulus.dyndns.org:6667")
		c.AddOption("Server", "nick", "testbot")
		c.AddOption("Server", "ident", "ident")
		c.AddOption("Server", "realname", "TestBot Client")
		c.AddOption("Server", "trigger", ".")
		c.WriteFile("go-faui2k11.cfg", 0644, "go-faui2k11 default config file")
		log.Println("Note: A new default configuration file has been generated in go-faui2k11.cfg. Please edit it to suit your needs and restart go-faui2k11 then")
		return
	}
	options := make(map[string]string)
	for _, x := range []string{"host", "nick", "ident", "realname"} {
		it, err := c.String("Server", x)
		if err != nil {
			log.Fatal(err)
			return
		}
		options[x] = it
	}
	var trigger byte
	strigger, err := c.String("Server", "trigger")
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(strigger) > 1 {
		log.Fatal("Trigger must be exactly one byte long")
	}

	s := ircclient.NewIRCClient(options["host"], options["nick"], options["realname"], options["ident"], trigger)
	ok = s.Connect()
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
