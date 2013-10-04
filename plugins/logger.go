package plugins

import (
	"../ircclient"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	default_logger_dir = "irclogs"
)

type LoggerPlugin struct {
	ic  *ircclient.IRCClient
}

func make_sure_dir_exists(dirname string) error {
	finfo, err := os.Lstat(dirname)
	if err != nil {
		return os.Mkdir(dirname, 0755) // hope the modes are ok like this
	}
	if finfo.IsDir() {
		return nil
	}
	// will fail, as file already exists, but this gives us clearer errors
	// than writing
	return os.Mkdir(dirname, 0755)
}

func (l *LoggerPlugin) Register(ic *ircclient.IRCClient) {
	l.ic = ic
	dir := l.ic.GetStringOption("Logger", "dir")
	if dir == "" {
		log.Println("added default logger dir value of \"" + default_logger_dir + "\" to config file")
		l.ic.SetStringOption("Logger", "dir", default_logger_dir)
		dir = default_logger_dir
	}
	// this is kind of an init function, let's check that stuff here
	err := make_sure_dir_exists(dir)
	if err != nil {
		log.Fatal(err)
	}
}

func (l *LoggerPlugin) String() string {
	return "logger"
}

func (l *LoggerPlugin) Info() string {
	return "logs ALL the irc"
}

func (l *LoggerPlugin) Usage(cmd string) string {
	// this method only exists for interface satisfaction
	// the logger plugin doesn't have any commands, so no
	// usage info is needed
	return ""
}

func (l *LoggerPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	return
}

func write_string_to_file(filename, msg string) error {
	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
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
			s = msg.Target
		} else { // query
			s = msg.Source
		}
		host := strings.SplitN(l.ic.GetStringOption("Server", "host"), ":", 2)[0]
		full_filename := l.ic.GetStringOption("Logger", "dir") + "/" + host + "_" + s
		msg := fmt.Sprintf("%s | %s: %s\n", time.Now().String(),
			strings.SplitN(msg.Source, "!", 2)[0], strings.Join(msg.Args, " "))
		if err := write_string_to_file(full_filename, msg); err != nil {
			log.Println(err.Error())
		}
	}
}

func (l *LoggerPlugin) Unregister() {
	return
}
