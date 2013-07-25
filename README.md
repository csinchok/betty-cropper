To run this on a mac:


    > brew install go
    > git clone git@gitlab.onion.com:csinchok/simpleimageserver.git
    > cd simpleimageserver
    > cp config.json.example config.json  # And edit the config accordingly...
    > export GOPATH=/Users/csinchok/Development/simpleimageserver/  # Change this to your project path.
    > go build -o server && ./server --root=/Users/csinchok/Development/simpleimageserver/config.json # Obviously change this path.

API:

POST an image (using the key "image") to localhost:8888/api/new, for example:

    > curl --form "image=@some_image.png" http://localhost:8888/api/new

This should return an image id ("1", if this is the first image). You can then visit http://localhost:8888/cropper/1, or an image URL, such as: http://localhost:8888/1/1x1/300.jpg


TODOs:

- Have the admin listen on a different interface.
- Memcached integration
- overall DRY'ing things up
- make ratios based on config file (currently hardcoded)