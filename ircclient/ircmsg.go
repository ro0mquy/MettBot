package ircclient

import (
	"strings"
	"log"
)

type IRCMessage struct {
	Source   string
	Target   string
	Command  string
	Args     []string
	Complete string
}

func ParseServerLine(line string) *IRCMessage {
	im := &IRCMessage{"", "", "", make([]string, 0), line}

	if len(line) == 0 || strings.Trim(line, " \t\n\r") == "" {
		log.Println("ParseIrcLine: empty line")
		return im
	}

	// source and target
	if line[0] == ':' {
		parts := strings.SplitN(line[1:], " ", 4) // 4: src cmd target rest
		im.Source = parts[0]
		im.Command = parts[1]
		im.Target = parts[2]
		// cut them off
		if len(parts) > 3 {
			line = parts[3]
		} else {
			line= ""
		}
	} else {
		parts := strings.SplitN(line, " ", 2) // cmd, rest
		im.Command = parts[0]
		if len(parts) > 1 {
			line= parts[1]
		} else {
			line= ""
		}
	}

	args := strings.SplitN(line, ":", 2)
	for _, a := range strings.Split(args[0], " ") {
		if a != "" {
			im.Args= append(im.Args, a)
		}
	}
	if len(args) > 1 {
		im.Args= append(im.Args, args[1])
	}

	//log.Printf("im: %#v\n", im)
	return im
}
