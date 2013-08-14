package main

import (
	"bytes"
	"compress/gzip"
	"io"
)

// jcrop_css returns raw, uncompressed file data.
func jcrop_css() []byte {
	gz, err := gzip.NewReader(bytes.NewBuffer([]byte{
0x1f,0x8b,0x08,0x00,0x00,0x09,0x6e,0x88,0x00,0xff,0xac,0x95,
0xc1,0x6e,0xa3,0x3c,0x10,0xc7,0xef,0x7d,0x0a,0x3e,0x55,0x9f,
0x94,0x54,0x31,0x75,0x92,0xed,0x46,0x0b,0xda,0x63,0x72,0xd8,
0xb7,0x30,0xe0,0x80,0x1b,0x63,0xb3,0xb6,0x49,0xd2,0xa2,0xbc,
0xfb,0xda,0x18,0x12,0x0c,0x4e,0xb5,0x91,0xf6,0xd0,0x06,0x7b,
0x86,0xff,0xcc,0xfc,0xc6,0x1e,0x5e,0x5f,0x82,0xf7,0xdf,0x35,
0x16,0x1f,0xe1,0xaf,0x54,0xf0,0x2a,0x2c,0x09,0x0b,0x53,0x29,
0x83,0x23,0x0c,0x7f,0x84,0xcb,0x55,0x30,0x4b,0x6a,0x42,0xb3,
0x68,0x05,0x97,0x6b,0xb8,0x5c,0x7d,0x9f,0x07,0x2f,0xaf,0x4f,
0xe1,0xbb,0x71,0x05,0x05,0xa7,0x19,0x16,0x4d,0x46,0x04,0x4e,
0x15,0xe1,0x2c,0xa2,0x4a,0xc4,0x0a,0x9f,0x15,0x40,0x94,0xe4,
0x7a,0x89,0xf7,0x2a,0xbe,0xf4,0xde,0x47,0x4a,0x18,0x5e,0xf4,
0xaf,0x9a,0x45,0x93,0xa0,0xf4,0x90,0x0b,0x5e,0xb3,0x2c,0x7a,
0xde,0xed,0x76,0x41,0x2d,0xe8,0xcc,0x66,0x91,0x93,0xfd,0x3c,
0xde,0x73,0xa6,0x80,0x24,0x9f,0x38,0x82,0x71,0xc5,0x25,0x69,
0x63,0xa0,0x44,0x72,0x5a,0x2b,0x3c,0x12,0x6e,0x0a,0x4c,0xf2,
0x42,0x45,0x4b,0x08,0xff,0x8f,0x4f,0x24,0x53,0x45,0xb4,0xac,
0xce,0xff,0x91,0xb2,0xe2,0x42,0x21,0x36,0xce,0x23,0x14,0xc6,
0xbb,0x69,0xff,0x6b,0xf5,0xab,0xb1,0x70,0xb4,0x1c,0x81,0x4e,
0xd4,0xe8,0xbb,0xee,0x61,0xc2,0x95,0xe2,0x65,0x63,0x7f,0x86,
0x6a,0x4a,0xe8,0x02,0x35,0x21,0x70,0xc2,0xc9,0x81,0x28,0xa0,
0x90,0x7e,0x45,0x4b,0x53,0x23,0x0f,0x52,0x4e,0xb9,0x88,0xb4,
0x0f,0x93,0x15,0x12,0x58,0x87,0xb8,0xfa,0xf1,0x3a,0x2d,0x40,
0x8a,0x28,0xe5,0xb5,0x8a,0x18,0x67,0xf8,0x6a,0xaa,0x25,0x16,
0x40,0x62,0xaa,0x89,0x5b,0x83,0xa7,0x6e,0x37,0x45,0xc4,0x32,
0x3a,0x44,0xdd,0xc5,0x7d,0x5e,0xaf,0xd7,0x71,0xc2,0x85,0xee,
0xa0,0x29,0x34,0x78,0xde,0x6e,0xb7,0x81,0x46,0x4b,0xb2,0x01,
0x77,0x6d,0xe8,0x03,0x6c,0xf4,0xa3,0xd5,0x37,0x4f,0x23,0xf9,
0x50,0xeb,0x00,0xd6,0x98,0x86,0x47,0x6f,0x3a,0x7a,0x89,0x44,
0x4e,0x18,0x68,0xd7,0xe0,0x9b,0xf6,0xef,0x36,0x14,0xaf,0xec,
0xda,0x3c,0x40,0xaf,0x8a,0xbc,0x71,0x1c,0xcb,0x75,0x86,0xa1,
0xe0,0x2d,0x82,0x4f,0x0b,0x37,0x9d,0x9b,0xed,0xb3,0x37,0x93,
0xfe,0x08,0x98,0x8d,0xb7,0x29,0xb8,0x56,0xe7,0x64,0x2b,0x83,
0x7f,0x57,0xd7,0x3d,0x15,0xf6,0x98,0x8c,0x1f,0x0f,0x7b,0xb4,
0xa6,0x3b,0x94,0xf1,0x0d,0xf3,0x7d,0xba,0x83,0x10,0xd3,0x9b,
0x32,0x10,0x3b,0x8d,0x7a,0xf6,0x95,0xa6,0xa7,0x63,0x99,0x40,
0x79,0x82,0x84,0xad,0x6f,0xe1,0xd9,0x94,0xcd,0xe4,0x10,0xba,
0x87,0x7c,0xe8,0x8c,0x7d,0x0a,0x27,0xcf,0x7c,0xd8,0xdc,0xcd,
0xa1,0x19,0x03,0xf5,0xba,0xc9,0x2f,0x11,0xfa,0x73,0xf3,0x34,
0x6f,0x42,0xd6,0xcd,0xfb,0x3e,0xb6,0x76,0x8a,0x04,0xbe,0xe9,
0xea,0x58,0xbc,0xa3,0x36,0xde,0x13,0xaa,0xf4,0xcd,0x47,0xb4,
0x2a,0xd0,0x8c,0x57,0x28,0x25,0xea,0xe3,0xe7,0x06,0xce,0x07,
0x03,0xaf,0xdb,0x8d,0xc2,0x0d,0xf4,0xcd,0x51,0x37,0x88,0x9d,
0x32,0xa0,0xe4,0x9f,0xc0,0x0e,0x15,0x20,0x50,0x46,0x6a,0x19,
0xad,0x75,0xca,0xfd,0xf0,0x9a,0x5a,0xa6,0x63,0x09,0x42,0xd8,
0x8d,0xa5,0x7e,0xc7,0xa4,0x3b,0x7d,0xf3,0x86,0x0b,0x89,0x83,
0x97,0xc2,0xd0,0x30,0x85,0x60,0xe2,0x3c,0x04,0xc1,0xc7,0xc0,
0x09,0xf1,0xcf,0x10,0x0c,0x0a,0x9e,0x42,0x71,0x11,0xb4,0x03,
0x1b,0x98,0xe2,0x46,0x08,0xa6,0x06,0xff,0x39,0xb8,0xb8,0x1f,
0xf3,0x80,0x94,0xf9,0x42,0xff,0x75,0x9b,0x95,0xc0,0x47,0x82,
0xcd,0x21,0x3c,0x03,0x7b,0x69,0xda,0x6f,0xce,0xe5,0xe9,0x4f,
0x00,0x00,0x00,0xff,0xff,0x68,0x06,0x64,0x0a,0x36,0x08,0x00,
0x00,
	}))

	if err != nil {
		panic("Decompression failed: " + err.Error())
	}

	var b bytes.Buffer
	io.Copy(&b, gz)
	gz.Close()

	return b.Bytes()
}