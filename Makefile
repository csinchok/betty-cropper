export GOPATH=$(realpath $(dir $(lastword $(MAKEFILE_LIST))))

all: reqs bindata build clean

reqs:
	@echo "\x1b[31;1mGetting dependencies...\x1b[0m"
	go get github.com/argusdusty/Ferret
	go get github.com/pmylund/go-cache
	go get github.com/rafikk/imagick/imagick
	go get github.com/csinchok/imgmin-go

build:
	@echo "\x1b[31;1mBuilding...\x1b[0m"
	go build

fulltests:
	@echo "\x1b[31;1mTesting...\x1b[0m"
	go test

runbench:
	go test -bench=CroppingJPEG

shorttests:
	go test --short

test: reqs fulltests clean

clean:
	@echo "\x1b[31;1mCleaning...\x1b[0m"
	rm -r testroot/*
	git checkout -- testroot
	cd testroot/1/ && ln -f -s Lenna.png src
	cd testroot/1234/5123 && ln -f -s Lemma.png src

bench: reqs runbench clean 