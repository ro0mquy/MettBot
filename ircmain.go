package main

import (
	"ircclient"
	"log"
	"plugins"
)

func main() {
	s := ircclient.NewIRCClient("go-faui2k11.cfg")
	s.RegisterPlugin(new(plugins.ListPlugins))
	s.RegisterPlugin(plugins.NewLoggerPlugin("irclogs"))
	s.RegisterPlugin(new(plugins.LecturePlugin))
	s.RegisterPlugin(new(plugins.QuitHandler))
	s.RegisterPlugin(new(plugins.QDevoicePlugin))
	s.RegisterPlugin(new(plugins.ChannelsPlugin))
	s.RegisterPlugin(new(plugins.AdminPlugin))
	ok := s.Connect()
	if ok != nil {
		log.Fatal(ok.String())
	}
	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.String())
	}
}
