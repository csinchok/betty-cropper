package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"image/draw"
	"image/color"
    "math"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"log"

	// "bettycropper/static"

	"github.com/disintegration/imaging"
	"code.google.com/p/freetype-go/freetype"
)

// TODOs: Shouldn't be opening the image file more than once.
// Memcached integration
// Admin interface on a different ip

var (
	configPath = flag.String("config", "/etc/betty-cropper/config.json", "Path for the config file")
	staticPath = flag.String("static", "/etc/betty-cropper/static/", "Path for the config file")
)

var imageRoot, adminListen, publicListen, publicAddress string  // Global config variables
var debug bool
var ratios []image.Point

var nextId = -1

func loadConfig() {
	if _, err := os.Stat(*staticPath); err != nil {
	    if os.IsNotExist(err) {
	    	workingDir, _ := os.Getwd()
	    	fmt.Printf("Static path \"%s\" doesn't exist. Using \"%s\" instead.\n", *staticPath, workingDir)
	        *staticPath = workingDir

	    }
	}

	if _, err := os.Stat(*configPath); err != nil {
		fmt.Printf("Can't read the config file at \"%s\", exiting.\n", *configPath)
		os.Exit(1)
	}


    type Config struct {
        ImageRoot string `json:"imageRoot"` // Where we put out images
        AdminListen string `json:"adminListen"` // The address that the admin API is served from
        PublicListen string `json:"publicListen"` // The address that the public interface is served from
        
        PublicAddress string `json:"publicAddress"` // The address that the public interface is served from
        Ratios []string `json:"ratios"` // A list of image ratios that we'll be cropping for
        Debug bool `json:"debug"` // If debug is true or false
    }

    var config Config
    if configPath != nil {
        configBytes, err := ioutil.ReadFile(*configPath)
        if err == nil {
            json.Unmarshal(configBytes, &config)
            adminListen = config.AdminListen
            publicListen = config.PublicListen
            publicAddress = config.PublicAddress
            imageRoot = config.ImageRoot

            debug = config.Debug

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
    adminListen = ":9999"
    publicListen = ":8888"
    publicAddress = "http://localhost:8888"
    imageRoot = "/var/betty-cropper"

}

func ratioStringToPoint(imageRatio string) image.Point {
    var w, _ = strconv.Atoi(strings.Split(imageRatio, "x")[0])
    var h, _ = strconv.Atoi(strings.Split(imageRatio, "x")[1])
    return image.Point{w, h}
}

func ratioPointToString(imageRatio image.Point) string {
	return ""
}


// TODO: Should imageRatio really be a string?
func getSelection(imageId string, imageSize image.Point, imageRatio string) image.Rectangle {

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
        aspect_ratio = ratioStringToPoint(imageRatio)
    }

	var original_ratio = float64(imageSize.X) / float64(imageSize.Y)
	var selection_ratio = float64(aspect_ratio.X) / float64(aspect_ratio.Y)

	var min = image.Pt(0, 0)
	var max = imageSize

	if selection_ratio < original_ratio {
		var x_offset = (float64(imageSize.X) - (float64(imageSize.Y) * float64(aspect_ratio.X) / float64(aspect_ratio.Y))) / 2.0
		min = image.Pt(int(math.Floor(x_offset)), 0)
		max = image.Pt(imageSize.X - int(math.Floor(x_offset)), imageSize.Y)
	}
	if selection_ratio > original_ratio {
		var y_offset = (float64(imageSize.Y) - (float64(imageSize.X) * float64(aspect_ratio.Y) / float64(aspect_ratio.X))) / 2.0

		min = image.Pt(0, int(math.Floor(y_offset)))
		max = image.Pt(imageSize.X, imageSize.Y - int(math.Floor(y_offset)))
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

	t := template.New("cropper.html")
	t.Parse(string(cropper_html()))
	t.Execute(w, map[string]interface{}{
		"ImageId":    imageId,
		"Ratios":     ratios,
		"Selections": selections,
		"ScaledSize": scaled_size,
        "ImageScale": imageScale,
        "PublicAddress": publicAddress,
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


func placeholder(w http.ResponseWriter, imageRatio string, width int, format string) {

	var ratio = ratioStringToPoint(imageRatio)
	var size = image.Rect(0, 0, width, int(math.Floor(float64(width) * float64(ratio.Y) / float64(ratio.X))))
	var dst = image.NewRGBA(size)
	backgroundColor := color.RGBA{204, 204, 204, 255}
	draw.Draw(dst, dst.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)

	fontBytes, err := ioutil.ReadFile("/Library/Fonts/Microsoft/Lucida Sans Unicode.ttf")
    if err == nil {
	    font, err := freetype.ParseFont(fontBytes)
	    if err != nil {
	        log.Println(err)
	        return
	    }

	    var txtImage = image.NewRGBA(image.Rect(0, 0, 600, 600))
	    draw.Draw(txtImage, txtImage.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)

	    darkGrey := image.NewUniform(color.RGBA{150, 150, 150, 255})

	    var fontSize = float64(width) * 52 / 300  // Stupid magic number

		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFont(font)
		c.SetFontSize(fontSize)
		c.SetClip(txtImage.Bounds())
		c.SetDst(txtImage)
		c.SetSrc(darkGrey)

		var offsetFix = int(math.Floor(fontSize * 12 / 72))  // Stupid magic number

		pt := freetype.Pt(0, int(c.PointToFix32(fontSize) >> 8) - offsetFix)
		pt, err = c.DrawString(imageRatio, pt)
	    if err != nil {
	        log.Println(err)
	        return
	    }

	    txtSize := image.Pt(int(pt.X >> 8), int(c.PointToFix32(fontSize) >> 8) + 2)

	    txtBounds := image.Rect(
	    	int(math.Floor(float64(size.Max.X) / 2.0) - (float64(txtSize.X) / 2.0)),
	    	int(math.Floor(float64(size.Max.Y) / 2.0) - (float64(txtSize.Y) / 2.0)),
	    	int(math.Floor(float64(size.Max.X) / 2.0) + (float64(txtSize.X) / 2.0)),
	    	int(math.Floor(float64(size.Max.Y) / 2.0) + (float64(txtSize.Y) / 2.0)),
	    )

	    draw.Draw(dst, txtBounds, txtImage, image.ZP, draw.Src)
    }

	if format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{100})
		return
	}
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
		return
	}

}


func crop(w http.ResponseWriter, r *http.Request) {

	var path_segments = strings.Split(r.URL.Path, "/")
	if len(path_segments) != 4 {
		http.Error(w, "Couldn't find that.", 404)
		return
	}

	var image_id = path_segments[1]
	var imageRatio = path_segments[2]
	var filename = path_segments[3]


    width, _ := strconv.Atoi(strings.Split(filename, ".")[0])
	if width > 9000 {
		http.Error(w, "Enhance your calm, bro.", 420)
		return
	}

	var format = strings.Split(filename, ".")[1]

    _, err := imaging.Open(imageRoot + "/" + image_id + "/src")
	if err != nil {
    	if debug {
    		placeholder(w, imageRatio, width, format)
    		return

    		
    	} else {
    		http.Error(w, "Couldn't find that.", 404)
    		return
    	}
    }

	var dst = imageCrop(image_id, imageRatio)
	dst = imaging.Resize(dst, int(width), 0, imaging.CatmullRom)

	_ = os.MkdirAll(imageRoot+"/"+image_id+"/"+imageRatio, 0700)
	outputWriter, _ := os.Create(imageRoot + r.URL.Path)

	if format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{90})
		jpeg.Encode(outputWriter, dst, &jpeg.Options{90})
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

	var adminServeMux = http.NewServeMux()
	var adminServer = http.Server{
		Addr: adminListen,
		Handler: adminServeMux,
	}

	var publicServeMux = http.NewServeMux()
	var publicServer = http.Server {
		Addr: publicListen,
		Handler: publicServeMux,
	}

	publicServeMux.HandleFunc("/", crop)
	go func() {
		publicServer.ListenAndServe()
	}()

	adminServeMux.Handle("/cropper/js/", http.StripPrefix("/cropper/js", http.FileServer(http.Dir(*staticPath + "/js"))))
	adminServeMux.Handle("/cropper/css/", http.StripPrefix("/cropper/css", http.FileServer(http.Dir(*staticPath + "/css"))))
	adminServeMux.HandleFunc("/cropper/", cropper)
	adminServeMux.HandleFunc("/api/new", newImage)
	adminServeMux.HandleFunc("/api/", api)
	adminServer.ListenAndServe()
	
}
