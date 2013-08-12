To quickly run this, setup a $GOPATH and $GOBIN, and then just go get the package. For example:

    > mkdir betty-test
    > cd betty-test
    > mkdir bin
    > export GOBIN=/Users/csinchok/Development/betty-test/bin
    > export GOPATH=/Users/csinchok/Development/betty-test/
    > go get github.com/disintegration/imaging
    > go get github.com/csinchok/betty-cropper
    > cp src/github.com/csinchok/betty-cropper/config.json.example ./config.json
    > mkdir /var/betty-cropper
    > ./bin/betty-cropper --config=/Users/csinchok/Development/betty-test/config.json --static=/Users/csinchok/Development/betty-test/src/github.com/csinchok/betty-cropper/

### API

POST an image (using the key "image") to localhost:8888/api/new, for example:

    > curl --form "image=@some_image.png" http://localhost:8888/api/new

This should return an image id ("1", if this is the first image). You can then visit http://localhost:8888/cropper/1, or an image URL, such as: http://localhost:8888/1/1x1/300.jpg

### To Develop on this:

    > brew install go
    > git clone git@gitlab.onion.com:csinchok/simpleimageserver.git
    > cd simpleimageserver
    > cp config.json.example config.json  # And edit the config accordingly...
    > mkdir -p /var/betty-cropper  # Or wherever you're gonna put your image root
    > export GOPATH=/Users/csinchok/Development/betty-cropper/  # Change this to your project path
    > go build -o betty-cropper && ./betty-cropper --config=/Users/csinchok/Development/simpleimageserver/config.json --static=/Users/csinchok/Development/simpleimageserver/ # Obviously change these paths

### TODOs

- Memcached integration
- overall DRY'ing things up
- Add JPEG quality option
- Remove cropped files after a selection changes.
- Allow image deletion