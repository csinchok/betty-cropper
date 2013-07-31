package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/disintegration/imaging"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
    "math"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// TODOs: Shouldn't be opening the image file more than once.
// Memcached integration
// Admin interface on a different ip
// CamelCase


var configPath = flag.String("config", "/etc/betty-cropper/config.json", "Path for the config file")

var staticPath = flag.String("static", "/etc/betty-cropper/static/", "Path for the config file")

var imageRoot, adminAddress, publicAddress string  // Global config variables
var ratios []image.Point

var nextId = -1

func loadConfig() {

    type Config struct {
        ImageRoot string // Where we put out images
        AdminAddress string // The address that the admin API is served from
        PublicAddress string // The address that the public interface is served from
        Ratios []string // A list of image ratios that we'll be cropping for
    }

    var config Config
    if configPath != nil {
        configBytes, err := ioutil.ReadFile(*configPath)
        if err == nil {
            json.Unmarshal(configBytes, &config)
            adminAddress = config.AdminAddress
            publicAddress = config.PublicAddress
            imageRoot = config.ImageRoot

            ratios = make([]image.Point, len(config.Ratios))
            // ratios = [config.Ratios.len()]image.Point
            for index,ratio := range config.Ratios {
            	var w, _ = strconv.Atoi(strings.Split(ratio, "x")[0])
            	var h, _ = strconv.Atoi(strings.Split(ratio, "x")[1])
            	ratios[index] = image.Pt(w, h)
            }
            return
        }
    }
    adminAddress = ":9999"
    publicAddress = ":8888"
    imageRoot = "/var/betty-cropper"

}


func getSelection(imageId string, image_size image.Point, imageRatio string) image.Rectangle {

    var selectionJsonPath = imageRoot + "/" + imageId + "/selections.json"
    var selections map[string]image.Rectangle
    selectionBytes, err := ioutil.ReadFile(selectionJsonPath)
    if err == nil {
        json.Unmarshal(selectionBytes, &selections)
    } else {
        // TODO: Make dynamic based on the number of ratios
        selections = make(map[string]image.Rectangle, 5)
    }
    // TODO: maybe pss the string into this?
    if selection, ok := selections[imageRatio]; ok {
        return selection
    }

    src, err := imaging.Open(imageRoot + "/" + imageId + "/src")
    if err != nil {
        fmt.Println("Couldn't find an image. Did you set the image root?")
    }

    var aspect_ratio = src.Bounds().Max
    if imageRatio != "original" {
        var w, _ = strconv.Atoi(strings.Split(imageRatio, "x")[0])
        var h, _ = strconv.Atoi(strings.Split(imageRatio, "x")[1])
        aspect_ratio = image.Point{w, h}
    }

	var original_ratio = float64(image_size.X) / float64(image_size.Y)
	var selection_ratio = float64(aspect_ratio.X) / float64(aspect_ratio.Y)

	var min = image.Pt(0, 0)
	var max = image_size

	if selection_ratio < original_ratio {
		var x_offset = (float64(image_size.X) - (float64(image_size.Y) * float64(aspect_ratio.X) / float64(aspect_ratio.Y))) / 2.0
		min = image.Pt(int(math.Floor(x_offset)), 0)
		max = image.Pt(image_size.X - int(math.Floor(x_offset)), image_size.Y)
	}
	if selection_ratio > original_ratio {
		var y_offset = (float64(image_size.Y) - (float64(image_size.X) * float64(aspect_ratio.Y) / float64(aspect_ratio.X))) / 2.0

		min = image.Pt(0, int(math.Floor(y_offset)))
		max = image.Pt(image_size.X, image_size.Y - int(math.Floor(y_offset)))
	}

	return image.Rectangle{min, max}
}

func imageCrop(imageId string, imageRatio string) image.Image {
	src, err := imaging.Open(imageRoot + "/" + imageId + "/src")
	if err != nil {
		fmt.Println("Couldn't find an image. Did you set the image root?")
	}

	var selection = getSelection(imageId, src.Bounds().Max, imageRatio)

	return imaging.Crop(src, selection)
}

func cropper(w http.ResponseWriter, r *http.Request) {
	var imageId = strings.Split(r.URL.Path, "/")[2]

	src, err := imaging.Open(imageRoot + "/" + imageId + "/src")
    var imageScale = 600.0 / float64(src.Bounds().Max.X)

	if err != nil {
		fmt.Println("Couldn't find an image. Did you set the image root?")
	}

	var selections = make([]image.Rectangle, len(ratios))

    // TODO: get selection from disk
	for i, ratio := range ratios {
        var imageRatio = fmt.Sprintf("%dx%d", ratio.X, ratio.Y)
		selections[i] = getSelection(imageId, src.Bounds().Max, imageRatio)
	}

	var scaled_size = image.Pt(600, int(600.0*float64(src.Bounds().Max.Y)/float64(src.Bounds().Max.X)))

	t, _ := template.ParseFiles(*staticPath + "/html/cropper.html")
	t.Execute(w, map[string]interface{}{
		"ImageId":    imageId,
		"Ratios":     ratios,
		"Selections": selections,
		"ScaledSize": scaled_size,
        "ImageScale": imageScale,
	})

	// t.Execute(w, Cropper{ImageId: image_id, Ratios: &ratios})
}

func api(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only, you asshole.", 405)
		return
	}

	var pathSegments = strings.Split(r.URL.Path, "/")
	var imageId = pathSegments[2]
	var imageRatio = pathSegments[3]

	minX, err := strconv.Atoi(r.FormValue("minX"))
    if err != nil {
        fmt.Println(err)
    }
	minY, err := strconv.Atoi(r.FormValue("minY"))
    if err != nil {
        fmt.Println(err)
    }

	maxX, err := strconv.Atoi(r.FormValue("maxX"))
    if err != nil {
        fmt.Println(err)
    }
	maxY, err := strconv.Atoi(r.FormValue("maxY"))
    if err != nil {
        fmt.Println(err)
    }

	var selections map[string]image.Rectangle

	var selection_json_path = imageRoot + "/" + imageId + "/selections.json"

	selection_bytes, err := ioutil.ReadFile(selection_json_path)
	if err == nil {
		json.Unmarshal(selection_bytes, &selections)
	} else {
        // TODO: Make dynamic based on the number of ratios
        selections = make(map[string]image.Rectangle, 5)
    }

	// TODO: validate image ratio
	selections[imageRatio] = image.Rectangle{
		image.Point{minX, minY},
		image.Point{maxX, maxY},
	}

	data, err := json.Marshal(selections)
	err = ioutil.WriteFile(selection_json_path, data, 0777)

	fmt.Fprintf(w, "Updated sucessfully")
}

func newImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only, you asshole.", 405)
		return
	}
	file, _, err := r.FormFile("image")
	// TODO: check to make sure it's a valid image
	if err != nil {
		http.Error(w, "POST error", 500)
		return
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "File error", 500)
	}

	var image_id = nextId
	nextId += 1

	_ = os.MkdirAll(imageRoot + "/" + strconv.Itoa(image_id), 0700)
	err = ioutil.WriteFile(imageRoot + "/" + strconv.Itoa(image_id) + "/src", data, 0777)
	if err != nil {
		http.Error(w, "IO error", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "{\"id\":%d}", image_id)
}

func crop(w http.ResponseWriter, r *http.Request) {

	var path_segments = strings.Split(r.URL.Path, "/")
	if len(path_segments) != 4 {
		http.Error(w, "Couldn't find that.", 404)
		return
	}

	var image_id = path_segments[1]
	var image_ratio = path_segments[2]
	var filename = path_segments[3]

	var dst = imageCrop(image_id, image_ratio)

    src, _ := imaging.Open(imageRoot + "/" + image_id + "/src")
    var width = src.Bounds().Max.X

    if strings.Split(filename, ".")[0] != "original" {
        width, _ = strconv.Atoi(strings.Split(filename, ".")[0])
    }
    
	if width > 9000 {
		http.Error(w, "Enhance your calm, bro.", 420)
		return
	}

	var format = strings.Split(filename, ".")[1]

	dst = imaging.Resize(dst, int(width), 0, imaging.CatmullRom)

	_ = os.MkdirAll(imageRoot+"/"+image_id+"/"+image_ratio, 0700)
	outputWriter, _ := os.Create(imageRoot + r.URL.Path)

	if format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{jpeg.DefaultQuality})
		jpeg.Encode(outputWriter, dst, &jpeg.Options{jpeg.DefaultQuality})
		return
	}
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
		png.Encode(outputWriter, dst)
		return
	}

	http.Error(w, "Couldn't find that.", 404)
	return
}

func main() {
	flag.Parse()

    loadConfig()

	fileList, _ := ioutil.ReadDir(imageRoot)
	if len(fileList) == 0 {
		nextId = 1
	} else {
		var lastDirectory = fileList[len(fileList)-1]
		lastId, _ := strconv.Atoi(lastDirectory.Name())
		nextId = lastId + 1
	}

	http.Handle("/cropper/js/", http.StripPrefix("/cropper/js", http.FileServer(http.Dir(*staticPath + "/js"))))
	http.Handle("/cropper/css/", http.StripPrefix("/cropper/css", http.FileServer(http.Dir(*staticPath + "/css"))))
	http.HandleFunc("/cropper/", cropper)
	http.HandleFunc("/api/new", newImage)
	http.HandleFunc("/api/", api)

	http.HandleFunc("/", crop)
	http.ListenAndServe(publicAddress, nil)
}
