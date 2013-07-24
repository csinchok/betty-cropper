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


var pwd, _ = os.Getwd()

var config_path = flag.String("config", "config.json", "Path for the config file")
var imageRoot, adminAddress, publicAddress string  // Global config variables

var next_id = -1

func loadConfig() {

    type Config struct {
        ImageRoot string // Where we put out images
        AdminAddress string // The address that the admin API is served from
        PublicAddress string // The address that the public interface is served from
        Ratios []string // A list of image ratios that we'll be cropping for
    }

    var config Config
    if config_path != nil {
        config_bytes, err := ioutil.ReadFile(*config_path)
        if err == nil {
            json.Unmarshal(config_bytes, &config)
            adminAddress = config.AdminAddress
            publicAddress = config.PublicAddress
            imageRoot = config.ImageRoot
        }
    }


}


func getSelection(image_id string, image_size image.Point, image_ratio string) image.Rectangle {

    var selection_json_path = imageRoot + "/" + image_id + "/selections.json"
    var selections map[string]image.Rectangle
    selection_bytes, err := ioutil.ReadFile(selection_json_path)
    if err == nil {
        json.Unmarshal(selection_bytes, &selections)
    } else {
        // TODO: Make dynamic based on the number of ratios
        selections = make(map[string]image.Rectangle, 5)
    }
    // TODO: maybe pss the string into this?
    if selection, ok := selections[image_ratio]; ok {
        return selection
    }

    src, err := imaging.Open(imageRoot + "/" + image_id + "/src")
    if err != nil {
        fmt.Println("Couldn't find an image. Did you set the image root?")
    }

    var aspect_ratio = src.Bounds().Max
    if image_ratio != "original" {
        var w, _ = strconv.Atoi(strings.Split(image_ratio, "x")[0])
        var h, _ = strconv.Atoi(strings.Split(image_ratio, "x")[1])
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

func imageCrop(image_id string, image_ratio string) image.Image {
	src, err := imaging.Open(imageRoot + "/" + image_id + "/src")
	if err != nil {
		fmt.Println("Couldn't find an image. Did you set the image root?")
	}

	var selection = getSelection(image_id, src.Bounds().Max, image_ratio)

	return imaging.Crop(src, selection)
}

func cropper(w http.ResponseWriter, r *http.Request) {
	var image_id = strings.Split(r.URL.Path, "/")[2]

	src, err := imaging.Open(imageRoot + "/" + image_id + "/src")
    var imageScale = 600.0 / float64(src.Bounds().Max.X)

	if err != nil {
		fmt.Println("Couldn't find an image. Did you set the image root?")
	}

	ratios := []image.Point{
		image.Point{1, 1},
		image.Point{2, 1},
		image.Point{3, 4},
		image.Point{4, 3},
		image.Point{16, 9},
	}

	var selections = make([]image.Rectangle, len(ratios))

    // TODO: get selection from disk
	for i, ratio := range ratios {
        var image_ratio = fmt.Sprintf("%dx%d", ratio.X, ratio.Y)
		selections[i] = getSelection(image_id, src.Bounds().Max, image_ratio)
	}

	var scaled_size = image.Pt(600, int(600.0*float64(src.Bounds().Max.Y)/float64(src.Bounds().Max.X)))

	t, _ := template.ParseFiles("cropper.html")
	t.Execute(w, map[string]interface{}{
		"ImageId":    image_id,
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

	var path_segments = strings.Split(r.URL.Path, "/")
	var image_id = path_segments[2]
	var image_ratio = path_segments[3]

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

	var selection_json_path = imageRoot + "/" + image_id + "/selections.json"

	selection_bytes, err := ioutil.ReadFile(selection_json_path)
	if err == nil {
		json.Unmarshal(selection_bytes, &selections)
	} else {
        // TODO: Make dynamic based on the number of ratios
        selections = make(map[string]image.Rectangle, 5)
    }

	// TODO: validate image ratio
	selections[image_ratio] = image.Rectangle{
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

	var image_id = next_id
	next_id += 1

	_ = os.MkdirAll(imageRoot+"/"+strconv.Itoa(image_id), 0700)
	err = ioutil.WriteFile(imageRoot+"/"+strconv.Itoa(image_id)+"/src", data, 0777)
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
	output_writer, _ := os.Create(imageRoot + r.URL.Path)

	if format == "jpg" {

		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{jpeg.DefaultQuality})
		jpeg.Encode(output_writer, dst, &jpeg.Options{jpeg.DefaultQuality})
		return
	}
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
		png.Encode(output_writer, dst)
		return
	}

	http.Error(w, "Couldn't find that.", 404)
	return
}

func main() {
	flag.Parse()

    loadConfig()

	file_list, _ := ioutil.ReadDir(imageRoot)
	if len(file_list) == 0 {
		next_id = 1
	} else {
		var last_directory = file_list[len(file_list)-1]
		last_id, _ := strconv.Atoi(last_directory.Name())
		next_id = last_id + 1
	}

	http.Handle("/cropper/js/", http.StripPrefix("/cropper/js", http.FileServer(http.Dir("./js"))))
	http.Handle("/cropper/css/", http.StripPrefix("/cropper/css", http.FileServer(http.Dir("./css"))))
	http.HandleFunc("/cropper/", cropper)
	http.HandleFunc("/api/new", newImage)
	http.HandleFunc("/api/", api)

	http.HandleFunc("/", crop)
	http.ListenAndServe(publicAddress, nil)
}
