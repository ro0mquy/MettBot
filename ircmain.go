package main

import (
	"ircclient"
	"log"
	"plugins"
	"time"
	"rand"
)

func main() {
	rand.Seed(time.Nanoseconds())
	s := ircclient.NewIRCClient("go-faui2k11.cfg")
	s.RegisterPlugin(new(plugins.KexecPlugin))
	s.RegisterPlugin(new(plugins.ListPlugins))
	s.RegisterPlugin(plugins.NewLoggerPlugin("irclogs"))
	s.RegisterPlugin(new(plugins.LecturePlugin))
	s.RegisterPlugin(new(plugins.QuitHandler))
	s.RegisterPlugin(new(plugins.QDevoicePlugin))
	s.RegisterPlugin(new(plugins.ChannelsPlugin))
	s.RegisterPlugin(new(plugins.AdminPlugin))
	s.RegisterPlugin(new(plugins.XKCDPlugin))
	s.RegisterPlugin(new(plugins.DecidePlugin))
	ok := s.Connect()
	if ok != nil {
		log.Fatal(ok.String())
	}
	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.String())
	}
}
