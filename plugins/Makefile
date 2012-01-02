GOBUILD = gobuild

all: plugins

clean:
	rm -rf *.o *.a *.[568vq] [568vq].out *.so _obj _test _testmain.go *.exe _cgo* test.out build.out

test:
	@echo "to be implemented"

plugins: $(wildcard *.go)
	make -C .. plugins

.PHONY:all test clean
