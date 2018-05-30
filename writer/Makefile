GOPATH=$(shell pwd)
export GOPATH

all: build

deps:
	go get -v -d .

build:
	go build -o main .

clean:
	rm -f main
