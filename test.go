package main

import (
	"ircclient"
	"fmt"
)

func main() {
	args := []string{"!das hier ist ein  \"Test! \\\"für\" das  "}
	msg := &ircclient.IRCMessage{"", "", "PRIVMSG", args, ""}
	ret := ircclient.ParseCommand(msg, '!')
	fmt.Printf("%#v", ret.Args)
}
