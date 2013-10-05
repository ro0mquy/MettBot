package plugins

import (
	"../ircclient"
	"log"
	"os"
	"strconv"
	"syscall"
)

type KexecPlugin struct {
	ic *ircclient.IRCClient
}

func (kp *KexecPlugin) Register(cl *ircclient.IRCClient) {
	kp.ic = cl
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
	socket := kp.ic.GetSocket()
	// check for error
	if socket == -1 {
		kp.ic.Reply(cmd, "Online restart failed")
		return
	}
	kp.ic.Reply(cmd, "Now trying online restart.")
	kp.ic.Shutdown()
	progname := os.Args[0]
	log.Println("kexec: " + progname)
	err := syscall.Exec(progname, []string{progname, strconv.Itoa(socket)}, []string{})
	// exec normally doesn't return
	kp.ic.Reply(cmd, "couldn't kexec: "+err.Error())
}

func (kp *KexecPlugin) Unregister() {
	// nothing to see here, move on
}
