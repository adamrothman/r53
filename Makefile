.PHONY : clean
BINARY = $(shell basename $(CURDIR))

build:
	go build

build-linux: GOOS := linux
build-linux: GOARCH := amd64
build-linux:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY)_$(GOOS)-$(GOARCH)

clean:
	-rm r53*
