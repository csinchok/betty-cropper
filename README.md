## Betty Cropper

[![Build Status](https://travis-ci.org/csinchok/betty-cropper.png?branch=master)](https://travis-ci.org/csinchok/betty-cropper)

### Get started:

    > cp config.json.example config.json  # Edit values as necessary.
    > make && ./betty-cropper --config=config.json

### API

__POST__ an image (using the key "image") to /api/new, for example:
    
    > curl --form "image=@Lenna.png" http://localhost:8698/api/new

This should return an image id ("1", if this is the first image).

You can get a cropped version of this image using a URL like: (http://localhost:8698/1/1x1/300.jpg)[http://localhost:8698/1/1x1/300.jpg].

To update the selections used for a crop, you can POST to /api/id/ratio, for example:

    > curl -d "maxX=511&maxY=511&minX=1&minY=1" http://localhost:8698/api/1/1x1

You can then visit http://localhost:8888/cropper/1, or an image URL, such as: 

__GET__ /api/search, with an option "q" parameter in order to get a list of files matching that description. For example:

    > curl -XGET http://localhost:8698/api/search?q=lenna

### TODOs

- Memcached integration
- overall DRY'ing things up
- Remove cropped files after a selection changes.
- Allow image deletion?