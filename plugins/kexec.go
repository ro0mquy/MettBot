package plugins

import (
	"ircclient"
	"syscall"
	"flag"
	"fmt"
)


type KexecPlugin struct {
	ic *ircclient.IRCClient
}

func (kp *KexecPlugin) Register(cl *ircclient.IRCClient) {
	kp.ic= cl
	kp.ic.RegisterCommandHandler("kexec", 0, 500, kp)
}

func (kp *KexecPlugin) String() string {
	return "kexec"
}

func (kp *KexecPlugin) Info() string {
	return "executes the bot from the bot without disconnect"
}

func (kp *KexecPlugin) Usage(cmd string) string {
	switch cmd {
	case "kexec":
		return "kexec: execute bot, thereby using new binary"
	}
	return ""
}

func (kp *KexecPlugin) ProcessLine(msg *ircclient.IRCMessage) {
	return
}

func (kp *KexecPlugin) ProcessCommand(cmd *ircclient.IRCCommand) {
	progname := flag.Arg(0)
	file, ferr := kp.ic.GetConn().File()
	if ferr != nil {
		kp.ic.Reply(cmd, "couldn't kexec: " + ferr.String())
		return
	}
	fd_arg_string := fmt.Sprintf("-fd=%d", file.Fd())
	err := syscall.Exec(progname, []string{progname, fd_arg_string}, []string{})
	// exec normally doesn't return
	kp.ic.Reply(cmd, "couldn't kexec: " + syscall.Errstr(err))
}

func (kp *KexecPlugin) Unregister() {
	// nothing to see here, move on
}

