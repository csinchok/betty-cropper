package main

import (
    // "log"
    "testing"
    "net/http"
)

func TestIdParsing(t *testing.T) {
    var request = ImageRequest{
        Id: "123",
        Ratio: "original",
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
    httpRequest, _ := http.NewRequest("GET", "/1234/16x9/600.jpg", nil)
    imageRequest, err := NewImageRequest(httpRequest) 
    if err != nil {
        t.Errorf("Request parsing error: %s", err)
    }

    if imageRequest.Id != "1234" {
        t.Errorf("Request parsing error (got '%s' for Id, should be '1234')", imageRequest.Id)
    }
    if imageRequest.Ratio != "16x9" {
        t.Errorf("Request parsing error (got '%s' for Ratio, should be '16x9')", imageRequest.Ratio)
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

}