GFLAGS+=-I $(realpath plugins) -I $(realpath ircclient)
GLFLAGS+=-L $(realpath plugins) -L $(realpath ircclient)
CLEANFILES=ircmain

ircmain:

ircmain.6: plugins/ ircclient/ 

plugins/: ircclient/

.DEFAULT: ircmain
