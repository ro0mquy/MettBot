package main

import (
	"ircclient"
	"log"
	"plugins"
)

func main() {
	confplugin := plugins.NewConfigPlugin()
	if confplugin == nil {
		return
	}
	c := confplugin.Conf
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
	if len(strigger) != 1 {
		log.Fatal("Trigger must be exactly one byte long")
	}
	trigger = strigger[0]

	s := ircclient.NewIRCClient(options["host"], options["nick"], options["realname"], options["ident"], trigger)
	s.RegisterPlugin(confplugin)
	s.RegisterPlugin(new(plugins.ListPlugins))
	s.RegisterPlugin(new(plugins.AuthPlugin))
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
	s.SendLine("JOIN #go-faui2k11") // For testing purposes
	ok = s.InputLoop()
	if ok != nil {
		log.Fatal(ok.String())
	}
}
