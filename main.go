package main

import (
	"./ircclient"
	"./plugins"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	s := ircclient.NewIRCClient("mettbot.cfg")
	s.RegisterPlugin(new(plugins.KexecPlugin))
	s.RegisterPlugin(new(plugins.ListPlugins))
	s.RegisterPlugin(new(plugins.LoggerPlugin))
	s.RegisterPlugin(new(plugins.QuitHandler))
	s.RegisterPlugin(new(plugins.ChannelsPlugin))
	s.RegisterPlugin(new(plugins.AdminPlugin))
	s.RegisterPlugin(new(plugins.TwitterPlugin))
	s.RegisterPlugin(new(plugins.DongPlugin))
	s.RegisterPlugin(new(plugins.TopicDiffPlugin))
	s.RegisterPlugin(new(plugins.MumblePlugin))

	ok := s.Connect()
	if ok != nil {
		log.Fatal(ok.Error())
	}

	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.Error())
	}
}
