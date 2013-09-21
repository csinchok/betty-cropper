export GOPATH=$(realpath $(dir $(lastword $(MAKEFILE_LIST))))

all: reqs bindata build clean

reqs:
	@echo "\x1b[31;1mGetting dependencies...\x1b[0m"
	go get github.com/argusdusty/Ferret
	go get code.google.com/p/freetype-go/freetype
	go get github.com/pmylund/go-cache
	go get github.com/jteeuwen/go-bindata
	go get github.com/gographics/imagick/imagick

bindata:
	@echo "\x1b[31;1mConverting static resources to golang...\x1b[0m"
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_droidsansmono_ttf.go" --func=font_ttf static/DroidSansMono.ttf

build:
	@echo "\x1b[31;1mBuilding...\x1b[0m"
	go build

testenv:
	@echo "\x1b[31;1mConfiguring test environment...\x1b[0m"
	rm -r testroot/*
	git checkout -- testroot
	cd testroot/1/ && ln -f -s Lenna.png src
	cd testroot/1234/5123 && ln -f -s Lemma.png src

fulltests:
	@echo "\x1b[31;1mTesting...\x1b[0m"
	go test

shorttests:
	go test --short

test: reqs bindata fulltests clean testenv

clean:
	@echo "\x1b[31;1mCleaning...\x1b[0m"
	rm -f ./bindata_*.go