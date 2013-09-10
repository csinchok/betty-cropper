### Setup:

    > cp config.json.example config.json  # Edit values as necessary.
    > make && ./betty-cropper

### API

__POST__ an image (using the key "image") to /api/new, for example:
    
    > curl --form "image=@some_image.png" http://localhost:8888/api/new

This should return an image id ("1", if this is the first image). You can then visit http://localhost:8888/cropper/1, or an image URL, such as: http://localhost:8888/1/1x1/300.jpg

__GET__ /api/search, with an option "q" parameter in order to get a list of files matching that description. For example:

    > curl -XGET http://localhost:8888/api/search?q=lenna

### TODOs

- Memcached integration
- overall DRY'ing things up
- Remove cropped files after a selection changes.
- Allow image deletion?