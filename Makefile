## warning: you need go-gb, which does all the work
## see http://code.google.com/p/go-gb/
## this Makefile only calls it apropriately

## find go-gb binary name..
ifneq (,$(shell which go-gb))
	GB = go-gb
else
	GB = gb
endif

GB_VERBOSE = -v

GB += $(GB_VERBOSE)

all: go-faui2k11

go-faui2k11:
	$(GB) .

clean:
	$(GB) -c .

nuke:
	$(GB) -N .

format:
	$(GB) --gofmt .

## these are only called by hand, actually go-gb does all the work
plugins:
	$(MAKE) -C plugins

plugins_test:
	$(MAKE) -C plugins test

ircclient:
	$(MAKE) -C ircclient

ircclient_test:
	$(MAKE) -C ircclient test

.PHONY: all
