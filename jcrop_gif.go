package main

import (
	"bytes"
	"compress/gzip"
	"io"
)

// jcrop_gif returns raw, uncompressed file data.
func jcrop_gif() []byte {
	gz, err := gzip.NewReader(bytes.NewBuffer([]byte{
0x1f,0x8b,0x08,0x00,0x00,0x09,0x6e,0x88,0x00,0xff,0x72,0xf7,
0x74,0xb3,0xb0,0x4c,0xe4,0x60,0xe0,0x60,0x98,0xc8,0xc0,0xb0,
0x6a,0xd5,0xaa,0xff,0xff,0xff,0x33,0x80,0x81,0xe2,0x7f,0x6e,
0x3f,0xd7,0x90,0x60,0x67,0xc7,0x00,0x57,0x23,0x3d,0x03,0x66,
0x46,0x90,0xd0,0x4f,0x16,0x4e,0x2e,0x20,0xad,0x03,0x92,0x07,
0x69,0x61,0x60,0xe2,0x9d,0xc2,0xca,0xb9,0x3c,0xfb,0x6c,0x88,
0x0f,0xfb,0xc5,0xc8,0xa7,0x3a,0xd8,0x54,0xf0,0xb7,0xe8,0x2d,
0x98,0x91,0xf5,0x3c,0x2e,0xfa,0xc2,0xc2,0x13,0x6f,0x96,0x6e,
0xc5,0x6a,0x0a,0x3f,0x4b,0x4b,0xd2,0xf4,0xd9,0xaf,0x2e,0x24,
0x5d,0xd8,0x12,0x91,0x59,0xbb,0x00,0xbb,0x9a,0x96,0xd6,0xa4,
0x09,0x55,0x3b,0x6f,0x25,0x65,0x1c,0x8a,0xd8,0x58,0xbd,0x02,
0xbb,0x6b,0x5a,0x14,0x35,0x09,0xb8,0x46,0x84,0x6f,0x11,0x41,
0xd7,0x4c,0x69,0x48,0x40,0xb8,0x66,0x11,0x58,0x0d,0x2b,0x86,
0x9a,0xc6,0x84,0x49,0xc8,0xae,0xb1,0x06,0x04,0x00,0x00,0xff,
0xff,0x90,0x9f,0x5d,0x35,0x49,0x01,0x00,0x00,
	}))

	if err != nil {
		panic("Decompression failed: " + err.Error())
	}

	var b bytes.Buffer
	io.Copy(&b, gz)
	gz.Close()

	return b.Bytes()
}