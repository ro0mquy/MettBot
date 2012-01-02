GOBUILD = gobuild

## subdirs is used for e.g. cleaning, testable_dirs lists only dirs
## which have a *_test.go file
SUBDIRS = ircclient plugins
TESTABLE_DIRS = ircclient



## default is ircmain
all: ircmain


## the plugins
plugins:
	##make -C plugins
	## geht im moment ned, komische include-pfad-sache
	gobuild -I . -lib=true plugins

## the ircclient library
ircclient:
	make -C ircclient

## the whole bot
#! wildcard: man könnte über alle SUBDIRS loopen und
#! sachen zamsuchen, weil man hier auch unrelated-zeug mitnimmt
ircmain: ircmain.go $(wildcard */*.go)
	$(GOBUILD) $<

## tests
test:
	@for dir in $(TESTABLE_DIRS) ; \
	do \
		pushd $${dir}; \
		echo "== testing $${dir} =="; \
		gotest; \
		popd; \
	done
	


## gobuild -clean=true fails
clean:
	rm -rf *.o *.a *.[568vq] [568vq].out *.so _obj _test _testmain.go *.exe _cgo* test.out build.out
	## clean subdirs
	@for dir in $(SUBDIRS); do make -C $${dir} clean; done

.PHONY: all test clean ircclient plugins
