package plugins

import (
	"ircclient"
	"os"
	"fmt"
	"strings"
	"time"
	"log"
)

type LoggerPlugin struct {
	ic  *ircclient.IRCClient
	dir string
}

func NewLoggerPlugin(_dir string) *LoggerPlugin {
	return &LoggerPlugin{dir: _dir}
}


func make_sure_dir_exists(dirname string) os.Error {
	finfo, err := os.Lstat(dirname)
	if err != nil {
		return os.Mkdir(dirname, 0755) // hope the modes are ok like this
	}
	if finfo.IsDirectory() {
		return nil
	}
	// will fail, as file already exists, but this gives us clearer errors
	// than writing 
	return os.Mkdir(dirname, 0755)
}

func (l *LoggerPlugin) Register(ic *ircclient.IRCClient) {
	l.ic= ic
	// this is kind of an init function, let's check that stuff here
	make_sure_dir_exists(l.dir)
}

func (l *LoggerPlugin) String() string {
	return "logger"
}

func (l *LoggerPlugin) Info() string {
	return "logs ALL the irc"
}

func (l *LoggerPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func write_string_to_file(filename, msg string) os.Error {
	fp, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	//fmt.Printf("%#v\n", fp)
	defer fp.Close()

	if _, err := fp.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}

func (l *LoggerPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	if msg.Command == "PRIVMSG" {
		var s string
		if msg.Target[0] == '#' { // channel
			s= msg.Target
		} else { // query
			s= msg.Source
		}
		host := strings.SplitN(l.ic.GetStringOption("Server", "hostport"), ":", 2)[0]
		full_filename := l.dir + "/" + host + "_" + s
		msg := fmt.Sprintf("%s | %s: %s\n", time.LocalTime().String(),
			strings.SplitN(msg.Source, "!", 2)[0],  strings.Join(msg.Args, " "))
		if err := write_string_to_file(full_filename, msg); err != nil {
			log.Println(err.String())
		}
	}
}

func (l *LoggerPlugin) Unregister() {
	return
}

