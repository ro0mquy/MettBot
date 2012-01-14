## actually only a go-gb wrapper

GB = go-gb

all: plugins


plugins:
	$(GB) .

test:
	$(GB) -t .

nuke:
	$(GB) -N .

format:
	$(GB) --gofmt .


.PHONY: all plugins test nuke format
