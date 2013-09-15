package main

import (
    // "log"
    "testing"
    "path/filepath"
    "net/http/httptest"
    // "net/http"
)

func TestIdParsing(t *testing.T) {
    var request = ImageRequest{
        Id: "123",
        RatioString: "original",
        Width: 600,
        Format: "jpg",
    }
    if request.Dir() != "/var/betty-cropper/123" {
        t.Errorf("Directory error (got '%s', should be '/var/betty-cropper/1234')", request.Dir())
    }
    if request.Path() != "/var/betty-cropper/123/original/600.jpg" {
        t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/original/600.jpg')", request.Path())
    }

    request.Id = "1234"
    if request.Dir() != "/var/betty-cropper/1234" {
        t.Errorf("Directory error (got '%s', should be '/var/betty-cropper/1234')", request.Dir())
    }
    if request.Path() != "/var/betty-cropper/1234/original/600.jpg" {
        t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/original/600.jpg')", request.Path())
    }

    request.Id = "12345"
    if request.Dir() != "/var/betty-cropper/1234/5" {
        t.Errorf("Directory error (got '%s', should be '/var/betty-cropper/1234/5')", request.Dir())
    }
    if request.Path() != "/var/betty-cropper/1234/5/original/600.jpg" {
        t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/5/original/600.jpg')", request.Path())
    }

    request.Id = "12345678"
    if request.Dir() != "/var/betty-cropper/1234/5678" {
        t.Errorf("Directory error (got '%s', should be '/var/betty-cropper/1234/5678')", request.Dir())
    }
    if request.Path() != "/var/betty-cropper/1234/5678/original/600.jpg" {
        t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/5678/original/600.jpg')", request.Path())
    }

    request.Id = "1234567890"
    if request.Dir() != "/var/betty-cropper/1234/5678/90" {
        t.Errorf("Directory error (got '%s', should be '/var/betty-cropper/1234/5678/90')", request.Dir())
    }
    if request.Path() != "/var/betty-cropper/1234/5678/90/original/600.jpg" {
        t.Errorf("Path error (got '%s', should be '/var/betty-cropper/1234/5678/90/original/600.jpg')", request.Path())
    }
}

func TestRequestParsing(t *testing.T) {
    // Test a standard request
    imageRequest, err := NewImageRequest("/1234/16x9/600.jpg") 
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
    if imageRequest.Dir() != "/var/betty-cropper/1234" {
        t.Errorf("Request parsing error (got '%s' for Dir(), should be '/var/betty-cropper/1234')", imageRequest.Dir())
    }
    if imageRequest.Path() != "/var/betty-cropper/1234/16x9/600.jpg" {
        t.Errorf("Request parsing error (got '%s' for Path(), should be '/var/betty-cropper/1234/16x9/600.jpg')", imageRequest.Path())
    }

    // Test a request with a larger Id
    imageRequest, err = NewImageRequest("/1234/567/16x9/600.jpg") 
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
    if imageRequest.Dir() != "/var/betty-cropper/1234/567" {
        t.Errorf("Request parsing error (got '%s' for Dir(), should be '/var/betty-cropper/1234/567')", imageRequest.Dir())
    }
    if imageRequest.Path() != "/var/betty-cropper/1234/567/16x9/600.jpg" {
        t.Errorf("Request parsing error (got '%s' for Path(), should be '/var/betty-cropper/1234/567/16x9/600.jpg')", imageRequest.Path())
    }

    // Make sure that bad requests fail...
    imageRequest, err = NewImageRequest("/12345/16x9/600.jpg") 
    if err == nil {
        t.Error("Request parsing error ('/12345/16x9/600.jpg' is an invalid URL, but didn't error)")
    }
    imageRequest, err = NewImageRequest("/1234/testing/600.jpg") 
    if err == nil {
        t.Error("Request parsing error ('/1234/testing/600.jpg' is an invalid URL, but didn't error)")
    }
    imageRequest, err = NewImageRequest("/1234/original/600.gif") 
    if err == nil {
        t.Error("Request parsing error ('/1234/original/600.gif' is an invalid URL, but didn't error)")
    }
}

func TestCrop(t *testing.T) {
    imageRoot, _ = filepath.Abs("testroot")

    resp := httptest.NewRecorder()
    _ = resp

}