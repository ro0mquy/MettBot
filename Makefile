include $(GOROOT)/src/Make.inc

TARG=ircclient
GOFILES=\
	ircclient/basicprotocol.go	\
	ircclient/ircclient.go		\
	ircclient/ircconn.go			\
	ircclient/ircmsg.go			\
	ircclient/plugin.go			\
	ircclient/throttle-ircu.go
include $(GOROOT)/src/Make.pkg
