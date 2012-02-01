GFLAGS+=-I $(realpath plugins) -I $(realpath ircclient)
GLFLAGS+=-L $(realpath plugins) -L $(realpath ircclient)
CLEANFILES=go-faui2k11 ircmain

go-faui2k11: ircmain
	cp ircmain go-faui2k11

ircmain:

ircmain.6: plugins/ ircclient/ 

plugins/: ircclient/

