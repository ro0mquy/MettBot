## actually only a go-gb wrapper

GB = go-gb

all: ircclient


ircclient:
	$(GB) .

test:
	$(GB) -t .

nuke:
	$(GB) -N .

format:
	$(GB) --gofmt .


.PHONY: all ircclient test nuke format
