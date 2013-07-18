package main

import (
	"flag"
    "strconv"
    "image"
	"image/png"
	"image/jpeg"
	// "image/gif"
	"strings"
    "net/http"
    "html/template"
    "os"
    "fmt"
    "imaging"
)

var pwd, _ = os.Getwd()
var image_root = flag.String("root", pwd, "The root of the image directory")


func imageCrop(image_id string, image_ratio string) image.Image {
    src, err := imaging.Open(*image_root + "/" + image_id + "/src")
    if(err != nil) {
        fmt.Println("Couldn't find an image. Did you set the image root?")
    }
    var aspect_width = src.Bounds().Max.X
    var aspect_height = src.Bounds().Max.Y
    if(image_ratio != "original") {
        aspect_width, _ = strconv.Atoi(strings.Split(image_ratio, "x")[0])
        aspect_height, _ = strconv.Atoi(strings.Split(image_ratio, "x")[1])
    }

    var original_ratio = float64(src.Bounds().Max.X) / float64(src.Bounds().Max.Y)
    var selection_ratio = float64(aspect_width) / float64(aspect_height)

    var min = image.Pt(0, 0)
    var max = src.Bounds().Max
    if(selection_ratio < original_ratio) {
        var x_offset = (float64(src.Bounds().Max.X) - (float64(src.Bounds().Max.Y) * float64(aspect_width) / float64(aspect_height))) / 2.0

        min = image.Pt(int(x_offset), 0)
        max = image.Pt(src.Bounds().Max.X - int(x_offset), src.Bounds().Max.Y)
    }
    if(selection_ratio > original_ratio) {
        var y_offset = (float64(src.Bounds().Max.Y) - (float64(src.Bounds().Max.X) * float64(aspect_height) / float64(aspect_width))) / 2.0

        min = image.Pt(0, int(y_offset))
        max = image.Pt(src.Bounds().Max.X, src.Bounds().Max.Y - int(y_offset))
    }

    var selection = image.Rectangle{min, max}

    return imaging.Crop(src, selection)
}

func admin(w http.ResponseWriter, r *http.Request) {

}

func cropper(w http.ResponseWriter, r *http.Request) {
    var image_id = strings.Split(r.URL.Path, "/")[2]
    type Cropper struct {
        ImageId string
    }


    t, _ := template.ParseFiles("cropper.html")
    t.Execute(w, Cropper{ImageId: image_id})
}

func api(w http.ResponseWriter, r *http.Request) {

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

    // TODO: write image to filesystem
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

    http.Handle("/cropper/js/", http.StripPrefix("/cropper/js", http.FileServer(http.Dir("./js"))))
    http.Handle("/cropper/css/", http.StripPrefix("/cropper/css", http.FileServer(http.Dir("./css"))))
    
    http.HandleFunc("/cropper/", cropper)
    http.HandleFunc("/api/", api)

    http.HandleFunc("/", handler)
    http.ListenAndServe(":8888", nil)
}