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
	s := ircclient.NewIRCClient("go-faui2k11.cfg")
	s.RegisterPlugin(new(plugins.KexecPlugin))
	s.RegisterPlugin(new(plugins.ListPlugins))
	s.RegisterPlugin(plugins.NewLoggerPlugin("irclogs"))
	//s.RegisterPlugin(new(plugins.LecturePlugin))
	s.RegisterPlugin(new(plugins.QuitHandler))
	s.RegisterPlugin(new(plugins.QDevoicePlugin))
	s.RegisterPlugin(new(plugins.ChannelsPlugin))
	s.RegisterPlugin(new(plugins.AdminPlugin))
	//s.RegisterPlugin(new(plugins.XKCDPlugin))
	s.RegisterPlugin(new(plugins.DecidePlugin))
	s.RegisterPlugin(new(plugins.EvaluationPlugin))
	ok := s.Connect()
	if ok != nil {
		log.Fatal(ok.Error())
	}
	// Has to be loaded after successful connection
	//s.RegisterPlugin(new(plugins.HalloWeltPlugin))
	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.Error())
	}
}
