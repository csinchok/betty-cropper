package main

import (
	// "log"
	"path/filepath"
	"testing"
	// "net/http/httptest"
	// "net/http"
)

func TestIdParsing(t *testing.T) {
	var request = BettyRequest{
		Id:          "123",
		RatioString: "original",
		Width:       600,
		Format:      "jpg",
	}
	if request.Path() != "/var/betty-cropper/123/original/600.jpg" {
		t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/original/600.jpg')", request.Path())
	}

	request.Id = "1234"
	if request.Path() != "/var/betty-cropper/1234/original/600.jpg" {
		t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/original/600.jpg')", request.Path())
	}

	request.Id = "12345"
	if request.Path() != "/var/betty-cropper/1234/5/original/600.jpg" {
		t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/5/original/600.jpg')", request.Path())
	}

	request.Id = "12345678"
	if request.Path() != "/var/betty-cropper/1234/5678/original/600.jpg" {
		t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/5678/original/600.jpg')", request.Path())
	}

	request.Id = "1234567890"
	if request.Path() != "/var/betty-cropper/1234/5678/90/original/600.jpg" {
		t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/5678/90/original/600.jpg')", request.Path())
	}
}

func TestRequestParsing(t *testing.T) {
	// Test a standard request
	imageRequest, err := ParseBettyRequest("/1234/16x9/600.jpg")
	if err != nil {
		t.Errorf("Request parsing error: %s", err)
	}

	if imageRequest.Id != "1234" {
		t.Errorf("Request parsing error (got '%s' for Id, should be '1234')", imageRequest.Id)
	}
	if imageRequest.RatioString != "16x9" {
		t.Errorf("Request parsing error (got '%s' for Ratio, should be '16x9')", imageRequest.RatioString)
	}
	if imageRequest.Width != 600 {
		t.Errorf("Request parsing error (got %d for width, should be 600)", imageRequest.Width)
	}
	if imageRequest.Format != "jpg" {
		t.Errorf("Request parsing error (got '%s' for format, should be 'jpg')", imageRequest.Format)
	}
	if imageRequest.Path() != "/var/betty-cropper/1234/16x9/600.jpg" {
		t.Errorf("Request parsing error (got '%s' for Path(), should be '/var/betty-cropper/1234/16x9/600.jpg')", imageRequest.Path())
	}

	// Test a request with a larger Id
	imageRequest, err = ParseBettyRequest("/1234/567/16x9/600.jpg")
	if err != nil {
		t.Errorf("Request parsing error: %s", err)
	}
	if imageRequest.Id != "1234567" {
		t.Errorf("Request parsing error (got '%s' for Id, should be '1234567')", imageRequest.Id)
	}
	if imageRequest.RatioString != "16x9" {
		t.Errorf("Request parsing error (got '%s' for Ratio, should be '16x9')", imageRequest.RatioString)
	}
	if imageRequest.Width != 600 {
		t.Errorf("Request parsing error (got %d for width, should be 600)", imageRequest.Width)
	}
	if imageRequest.Format != "jpg" {
		t.Errorf("Request parsing error (got '%s' for format, should be 'jpg')", imageRequest.Format)
	}
	if imageRequest.Path() != "/var/betty-cropper/1234/567/16x9/600.jpg" {
		t.Errorf("Request parsing error (got '%s' for Path(), should be '/var/betty-cropper/1234/567/16x9/600.jpg')", imageRequest.Path())
	}

	// Make sure that bad requests fail...
	imageRequest, err = ParseBettyRequest("/12345/16x9/600.jpg")
	if err == nil {
		t.Error("Request parsing error ('/12345/16x9/600.jpg' is an invalid URL, but didn't error)")
	}
	imageRequest, err = ParseBettyRequest("/1234/testing/600.jpg")
	if err == nil {
		t.Error("Request parsing error ('/1234/testing/600.jpg' is an invalid URL, but didn't error)")
	}
	imageRequest, err = ParseBettyRequest("/1234/original/600.gif")
	if err == nil {
		t.Error("Request parsing error ('/1234/original/600.gif' is an invalid URL, but didn't error)")
	}
}

func TestBettyImage(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }

	imageRoot, _ = filepath.Abs("testroot")
	debug = false

	// Test with a short id
	img, err := GetBettyImage("1")
	if err != nil {
		t.Errorf("Error getting image info: %s", err.Error())
	}
	if img.Filename != "Lenna.png" {
		t.Errorf("Filename should be 'Lenna.png', but we got '%s'", img.Filename)
	}
	if img.Size.X != 512 || img.Size.Y != 512 {
		t.Errorf("Size should be '512x512', but we got '%dx%d'", img.Size.X, img.Size.Y)
	}

	// Test with a longer id
	img, err = GetBettyImage("12345123")
	if err != nil {
		t.Errorf("Error getting image info: %s", err.Error())
	}
	if img.Filename != "Lenna.png" {
		t.Errorf("Filename should be 'Lenna.png', but we got '%s'", img.Filename)
	}
	if img.Size.X != 512 || img.Size.Y != 512 {
		t.Errorf("Size should be '512x512', but we got '%dx%d'", img.Size.X, img.Size.Y)
	}
}
