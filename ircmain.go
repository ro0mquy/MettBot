package main

import (
	"ircclient"
	"log"
	"plugins"
)

func main() {
	/* TODO: Move that to config plugin
	options := make(map[string]string)
	for _, x := range []string{"host", "nick", "ident", "realname"} {
		it, err := c.String("Server", x)
		if err != nil {
			log.Fatal(err)
			return
		}
		options[x] = it
	}
	trigger, err := c.String("Server", "trigger")
	if err != nil {
		log.Fatal(err)
		return
	}
	if utf8.RuneCountInString(trigger) != 1 {
		log.Fatal("Trigger must be exactly one unicode rune long")
	}
	*/

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
