GC=6g
GFLAGS+=
GL=6l
GLFLAGS+=
GOFILES=$(wildcard *.go)
DIRS=$(wildcard */)
CLEANFILES=

export GFLAGS
export GLFLAGS
export CLEANFILES

.PRECIOUS: %.6
.PHONY: FORCE clean

include $(wildcard Makefile)

$(DIRS) : FORCE
	@echo "Building package $@"
	@cd "$@" && $(MAKE) -f "$(realpath $(@D))/../makefile" "$(@:/=).6"


$(GOFILES:.go=.6) : $(GOFILES) 
	$(GC) $(GFLAGS) -o $@ $(wildcard *.go)


info:
	@echo ""
	@echo "Plugins: $(DIRS)"
	@echo "BaseFiles: $(GOFILES)"
	@echo "GFLAGS: $(GFLAGS)"
	@echo "GLFLAGS: $(GLFLAGS)"

FORCE:

%.6   : $(wildcard *.go)
	$(GC) $(GFLAGS) -o $@ $(wildcard *.go)

%:: %.6
	$(GL) $(GLFLAGS) -o $@ $?

clean:
	@echo "Removing files"
	@for f in $(DIRS); do rm -f $$f$${f%%/}.6; done
	@for f in $(CLEANFILES); do rm -f $$f; done
	@rm -f *.6

