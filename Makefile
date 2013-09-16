export GOPATH=$(realpath $(dir $(lastword $(MAKEFILE_LIST))))

all: reqs bindata build clean

reqs:
	@echo "\x1b[31;1mGetting dependencies...\x1b[0m"
	go get github.com/argusdusty/Ferret
	go get code.google.com/p/freetype-go/freetype
	go get github.com/disintegration/imaging
	go get github.com/pmylund/go-cache
	go get github.com/jteeuwen/go-bindata

bindata:
	@echo "\x1b[31;1mConverting static resources to golang...\x1b[0m"
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_droidsansmono_ttf.go" --func=font_ttf static/DroidSansMono.ttf
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_jcrop_gif.go" --func=css_jcrop_gif static/Jcrop.gif
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_jquery_jcrop_min_css.go" --func=css_jquery_jcrop_min_css static/jquery.Jcrop.min.css
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_jquery_color_js.go" --func=js_jquery_color_js static/jquery.color.js
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_jquery_jcrop_min_js.go" --func=js_jquery_jcrop_min_js static/jquery.Jcrop.min.js
	$(GOPATH)/bin/go-bindata --out="$(GOPATH)/bindata_cropper_html.go" --func=html_cropper_html static/cropper.html

build:
	@echo "\x1b[31;1mBuilding...\x1b[0m"
	go build

link:
	cd testroot/1/ && ln -f -s Lenna.png src
	cd testroot/1234/5123 && ln -f -s Lemma.png src

fulltests:
	@echo "\x1b[31;1mTesting...\x1b[0m"
	go test

shorttests:
	go test --short

test: reqs bindata link fulltests clean

clean:
	@echo "\x1b[31;1mCleaning...\x1b[0m"
	rm -f ./bindata_*.go