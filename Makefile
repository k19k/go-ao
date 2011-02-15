include $(GOROOT)/src/Make.inc

TARG=ao
CGOFILES=ao.go
CGO_LDFLAGS=-lao

include $(GOROOT)/src/Make.pkg
