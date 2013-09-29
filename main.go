package main

import (
	"./ircclient"
	"log"
	"math/rand"
	"./plugins"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	s := ircclient.NewIRCClient("mettbot.cfg")
	s.RegisterPlugin(new(plugins.KexecPlugin))
	s.RegisterPlugin(new(plugins.ListPlugins))
	s.RegisterPlugin(plugins.NewLoggerPlugin("irclogs"))
	s.RegisterPlugin(new(plugins.QuitHandler))
	s.RegisterPlugin(new(plugins.ChannelsPlugin))
	s.RegisterPlugin(new(plugins.AdminPlugin))
	s.RegisterPlugin(new(plugins.TwitterPlugin))
	s.RegisterPlugin(new(plugins.DongPlugin))
	s.RegisterPlugin(new(plugins.TopicDiffPlugin))

	ok := s.Connect()
	if ok != nil {
		log.Fatal(ok.Error())
	}

	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.Error())
	}
}
