package main

import (
    "fmt"
    // "log"
    "image"
    "io/ioutil"
    "os"
	"path/filepath"
	"testing"
	"net/http/httptest"
	"net/http"
)

func TestIdParsing(t *testing.T) {
    config.ImageRoot = "/var/betty-cropper"
    config.ImgminEnabled = false
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
    config.ImgminEnabled = false
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
	config.ImageRoot, _ = filepath.Abs("testroot")
	config.PlaceholderEnabled = false
    config.ImgminEnabled = false

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
    if img.Selections["3x1"] != image.Rect(0, 144, 512, 314) {
        t.Errorf("Selection['3x1'] should be '0,144,512,314', but it's not.")
    }

	// Test with a longer id
	img, err = GetBettyImage("12345123")
	if err != nil {
		t.Errorf("Error getting image info: %s", err.Error())
	}
	if img.Filename != "Lemma.png" {
		t.Errorf("Filename should be 'Lemma.png', but we got '%s'", img.Filename)
	}
	if img.Size.X != 512 || img.Size.Y != 512 {
		t.Errorf("Size should be '512x512', but we got '%dx%d'", img.Size.X, img.Size.Y)
	}
}

func TestIndexing(t *testing.T) {
    config.ImageRoot, _ = filepath.Abs("testroot")
    config.PlaceholderEnabled = false
    config.ImgminEnabled = false

    buildIndex()

    if nextId != 12345124 {
        t.Errorf("nextId should be 12345124, but is %d", nextId)
    }

    ids, _ := SearchEngine.Query("", 25)
    if len(ids) != 2 {
        t.Errorf("Found %d results for '', there should be 2", len(ids))
    }

    ids, _ = SearchEngine.Query("lemma", 25)
    if len(ids) != 1 {
        t.Errorf("Found %d results for 'lemma', there should be 1", len(ids))
    }

    ids, _ = SearchEngine.Query("lenna", 25)
    if len(ids) != 1 {
        t.Errorf("Found %d results for 'lenna', there should be 1", len(ids))
    }
}

func TestSetters(t *testing.T) {
    config.ImgminEnabled = false
    img, err := GetBettyImage("1")
    if err != nil {
        t.Errorf("Error getting image info: %s", err.Error())
    }

    // Test the Setters
    err = img.SetSelection("3x1", image.Rect(0, 140, 512, 310))
    if err != nil {
        t.Errorf("Error setting image selection: %s", err.Error())
    }
    if img.Selections["3x1"] != image.Rect(0, 140, 512, 310) {
        t.Errorf("Selection['3x1'] should be '0,140,512,310', but it's not.")
    }
    err = img.SetName("Farts")
    if err != nil {
        t.Errorf("Error setting image name: %s", err.Error())
    }
    if img.Filename != "Farts.png" {
        t.Errorf("Filename should be 'Farts.png', but we got '%s'", img.Filename)
    }
    err = img.SetCredit("Farty McFarter")
    if err != nil {
        t.Errorf("Error setting image credit: %s", err.Error())
    }
    if img.Credit != "Farty McFarter" {
        t.Errorf("Credit should be 'Farty McFarter', but we got '%s'", img.Credit)
    }
}

func TestCropping(t *testing.T) {
    config.ImageRoot, _ = filepath.Abs("testroot")
    config.PlaceholderEnabled = false
    config.ImgminEnabled = false

    server := httptest.NewServer(http.HandlerFunc(crop))

    uri := "/1/16x9/200.jpg"
    resp, err := http.Get(server.URL + uri)
    if err != nil {
        t.Error(err.Error())
    }

    if _, err := ioutil.ReadAll(resp.Body); err != nil {
        t.Fail()
    } else {
        _, err := os.Stat(filepath.Join(config.ImageRoot, "/1/16x9/200.jpg"))
        if err != nil && os.IsNotExist(err) {
            t.Error("Didn't create crop")
        }
    }
}

func BenchmarkCroppingJPEG(b *testing.B) {

    ratioStrings := []string{"1x1", "2x1", "3x1", "3x4", "4x3", "16x9"}
    config.ImageRoot, _ = filepath.Abs("testroot")
    config.PlaceholderEnabled = false
    config.ImgminEnabled = false

    server := httptest.NewServer(http.HandlerFunc(crop))

    for i := 0; i < b.N; i++ {
        ratio := ratioStrings[i % len(ratioStrings)]
        uri := fmt.Sprintf("/1/%s/%d.jpg", ratio, 100 + (i % 2000))
        resp, err := http.Get(server.URL + uri)
        if err != nil {
            b.Error(err.Error())
        }
        _ = resp
        //  _ = resp.Body
        // if err != nil {
        //     log.Println(err.Error())
        // }
    }
}
