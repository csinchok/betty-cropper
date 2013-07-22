package main

import (
	"flag"
    "strconv"
    "image"
	"image/png"
	"image/jpeg"
    "io/ioutil"
	"strings"
    "net/http"
    "html/template"
    "os"
    "fmt"
    "github.com/disintegration/imaging"
)

var pwd, _ = os.Getwd()
var image_root = flag.String("root", pwd, "The root of the image directory")
var next_id = -1

func defaultSelection(image_size image.Point, aspect_ratio image.Point) image.Rectangle {

    var original_ratio = float64(image_size.X) / float64(image_size.Y)
    var selection_ratio = float64(aspect_ratio.X) / float64(aspect_ratio.Y)

    var min = image.Pt(0, 0)
    var max = image_size

    if(selection_ratio < original_ratio) {
        var x_offset = (float64(image_size.X) - (float64(image_size.Y) * float64(aspect_ratio.X) / float64(aspect_ratio.Y))) / 2.0
        min = image.Pt(int(x_offset), 0)
        max = image.Pt(image_size.X - int(x_offset), image_size.Y)
    }
    if(selection_ratio > original_ratio) {
        var y_offset = (float64(image_size.Y) - (float64(image_size.X) * float64(aspect_ratio.Y) / float64(aspect_ratio.X))) / 2.0

        min = image.Pt(0, int(y_offset))
        max = image.Pt(image_size.X, image_size.Y - int(y_offset))
    }

    return image.Rectangle{min, max}
}


func imageCrop(image_id string, image_ratio string) image.Image {
    src, err := imaging.Open(*image_root + "/" + image_id + "/src")
    if(err != nil) {
        fmt.Println("Couldn't find an image. Did you set the image root?")
    }

    var aspect_ratio = src.Bounds().Max
    if(image_ratio != "original") {
        var w, _ = strconv.Atoi(strings.Split(image_ratio, "x")[0])
        var h, _ = strconv.Atoi(strings.Split(image_ratio, "x")[1])
        aspect_ratio = image.Point{w, h}
    }

    var selection = defaultSelection(src.Bounds().Max, aspect_ratio)

    return imaging.Crop(src, selection)
}

func cropper(w http.ResponseWriter, r *http.Request) {
    var image_id = strings.Split(r.URL.Path, "/")[2]

    src, err := imaging.Open(*image_root + "/" + image_id + "/src")
    if(err != nil) {
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

    for i, ratio := range ratios {
        selections[i] = defaultSelection(src.Bounds().Max, ratio);
    }

    var scaled_size = image.Pt(600, int(600.0 * float64(src.Bounds().Max.Y) / float64(src.Bounds().Max.X)));

    t, _ := template.ParseFiles("cropper.html")
    t.Execute(w, map[string]interface{} {
        "ImageId": image_id,
        "Ratios": ratios,
        "Selections":  selections,
        "Size": src.Bounds().Max,
        "ScaledSize": scaled_size,
    })

    // t.Execute(w, Cropper{ImageId: image_id, Ratios: &ratios})
}

func api(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/api/new" {

    }
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
    
    _ = os.MkdirAll(*image_root + "/" + strconv.Itoa(image_id), 0700)
    err = ioutil.WriteFile(*image_root + "/" + strconv.Itoa(image_id) + "/src", data, 0777) 
    if err != nil { 
        http.Error(w, "IO error", 500)
    }
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, "{\"id\":%d}", image_id)
}


func handler(w http.ResponseWriter, r *http.Request) {
	
 	var path_segments = strings.Split(r.URL.Path, "/")
 	if len(path_segments) != 4 {
 		http.Error(w, "Couldn't find that.", 404)
        return
 	}
 	
    var image_id = path_segments[1] 	
 	var image_ratio = path_segments[2]
    var filename = path_segments[3]

    var dst = imageCrop(image_id, image_ratio)
    
    var width, _ = strconv.Atoi(strings.Split(filename, ".")[0])
    if (width > 5000) {
        http.Error(w, "Enhance your calm, bro.", 420)
        return
    }

    var format = strings.Split(filename, ".")[1]
    _ = format

    dst = imaging.Resize(dst, int(width), 0, imaging.CatmullRom)

    _ = os.MkdirAll(*image_root + "/" + image_id + "/" + image_ratio, 0700)
    output_writer, _ := os.Create(*image_root + r.URL.Path)

    if (format == "jpg") {
        
        w.Header().Set("Content-Type", "image/jpeg")
        jpeg.Encode(w, dst, &jpeg.Options{jpeg.DefaultQuality})
        jpeg.Encode(output_writer, dst, &jpeg.Options{jpeg.DefaultQuality})
        return
    }
    if (format == "png") {
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

    file_list, _ := ioutil.ReadDir(*image_root)
    var last_directory = file_list[len(file_list) - 1]
    last_id, _ := strconv.Atoi(last_directory.Name())
    next_id = last_id + 1

    http.Handle("/cropper/js/", http.StripPrefix("/cropper/js", http.FileServer(http.Dir("./js"))))
    http.Handle("/cropper/css/", http.StripPrefix("/cropper/css", http.FileServer(http.Dir("./css"))))
    
    http.HandleFunc("/cropper/", cropper)
    http.HandleFunc("/api/new", newImage)
    http.HandleFunc("/api/", api)

    http.HandleFunc("/", handler)
    http.ListenAndServe(":8888", nil)
}