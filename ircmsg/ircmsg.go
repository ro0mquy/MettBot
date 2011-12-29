package ircmsg


import (
	"strings"
	"log"
)

type IRCMessage struct {
	Source string
	Target string
	Command string
	Args []string
	Complete string
}


func ParseServerLine(line string) *IRCMessage {
	im := &IRCMessage{"", "", "", make([]string, 0), line}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		log.Println("ParseIrcLine: empty line")
		return im
	}


	// source and target
	if parts[0][0] == ':' {
		im.Source= strings.Replace(parts[0], ":", "", 1)
		im.Command= parts[1]
		im.Target= parts[2]
		// cut them off
		parts= parts[3:]
		line= strings.Replace(line, ":", "", 1)
	} else {
		im.Command= parts[0]
		parts= parts[1:]
	}

	for _, s := range parts {
		if s[0] == ':' {
			break
		}
		im.Args= append(im.Args, s)
	}
	// line has the leading ':' cut off, if it was ever there
	// (see "source and target" above)
	lastargpos:= strings.Index(line, ":")
	im.Args= append(im.Args, line[lastargpos+1:])

	log.Printf("im: %#v\n", im)
	return im
}
