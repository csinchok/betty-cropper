package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"code.google.com/p/freetype-go/freetype"
	"github.com/disintegration/imaging"
)

var BETTY_VERSION = "1.1.15"

// TODOs: Shouldn't be opening the image file more than once.
// Memcached integration
// Admin interface on a different ip

var (
	version       = flag.Bool("version", false, "Print the version number and exit")
	configPath    = flag.String("config", "config.json", "Path for the config file")
	imageRoot     = "/var/betty-cropper"
	adminListen   = ":9999"
	publicListen  = ":8888"
	publicAddress = "localhost:8888"
	debug         = false
	imgmin        = false
	ratios        []image.Point
	nextId        = -1
	adminReady    = false
)

func loadConfig() {

	if *version {
		fmt.Printf(BETTY_VERSION)
		os.Exit(0)
	}

	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		log.Printf("\"%s\" is a bad path for a config file, exiting.\n", *configPath)
		os.Exit(1)
	}

	_, err = exec.LookPath("imgmin")
	if err != nil {
		log.Println("Couldn't find imgmin in the $PATH, compression won't be as effective.")
	} else {
		imgmin = true
	}

	type Config struct {
		ImageRoot    string `json:"imageRoot"`    // Where we put out images
		AdminListen  string `json:"adminListen"`  // The address that the admin API is served from
		PublicListen string `json:"publicListen"` // The address that the public interface is served from

		PublicAddress string   `json:"publicAddress"` // The address that the public interface is served from
		Ratios        []string `json:"ratios"`        // A list of image ratios that we'll be cropping for
		Debug         bool     `json:"debug"`         // If debug is true or false
	}

	var config Config
	configBytes, err := ioutil.ReadFile(absConfigPath)
	if err != nil {
        log.Printf("Can't read the config file, because \"%s\", exiting.\n", err)
        os.Exit(1)
    }
	json.Unmarshal(configBytes, &config)
	adminListen = config.AdminListen
	publicListen = config.PublicListen
	publicAddress = config.PublicAddress
	imageRoot = config.ImageRoot

	debug = config.Debug

	ratios = make([]image.Point, len(config.Ratios))
	// ratios = [config.Ratios.len()]image.Point
	for index, ratio := range config.Ratios {
		var w, _ = strconv.Atoi(strings.Split(ratio, "x")[0])
		var h, _ = strconv.Atoi(strings.Split(ratio, "x")[1])
		ratios[index] = image.Pt(w, h)
	}
}

func ratioStringToPoint(imageRatio string) image.Point {
	var w, _ = strconv.Atoi(strings.Split(imageRatio, "x")[0])
	var h, _ = strconv.Atoi(strings.Split(imageRatio, "x")[1])
	return image.Point{w, h}
}

func getSelection(imageId string, imageSize image.Point, imageRatio string) image.Rectangle {

	var selectionJsonPath = imageRoot + "/" + imageId + "/selections.json"
	var selections map[string]image.Rectangle
	selectionBytes, err := ioutil.ReadFile(selectionJsonPath)
	if err == nil {
		json.Unmarshal(selectionBytes, &selections)
	} else {
		// TODO: Make dynamic based on the number of ratios
		selections = make(map[string]image.Rectangle, len(ratios))
	}
	// TODO: maybe pss the string into this?
	if selection, ok := selections[imageRatio]; ok {
		return selection
	}

	src, err := imaging.Open(filepath.Join(imageRoot, imageId, "src"))
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
		max = image.Pt(imageSize.X-int(math.Floor(x_offset)), imageSize.Y)
	}
	if selection_ratio > original_ratio {
		var y_offset = (float64(imageSize.Y) - (float64(imageSize.X) * float64(aspect_ratio.Y) / float64(aspect_ratio.X))) / 2.0

		min = image.Pt(0, int(math.Floor(y_offset)))
		max = image.Pt(imageSize.X, imageSize.Y-int(math.Floor(y_offset)))
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

var backgroundColors = []color.RGBA{
	color.RGBA{153, 153, 51, 255},
	color.RGBA{102, 153, 51, 255},
	color.RGBA{51, 153, 51, 255},
	color.RGBA{153, 51, 51, 255},
	color.RGBA{194, 133, 71, 255},
	color.RGBA{51, 153, 102, 255},
	color.RGBA{153, 51, 102, 255},
	color.RGBA{71, 133, 194, 255},
	color.RGBA{51, 153, 153, 255},
	color.RGBA{153, 51, 153, 255},
}
var backgroundIndex = 0

func placeholder(w http.ResponseWriter, imageReq ImageRequest) {
	// TODO: Don't do so much stupid shit with this font stuff.

	var ratio = ratioStringToPoint(imageReq.ratio)
	var size = image.Rect(0, 0, imageReq.width, int(math.Floor(float64(imageReq.width)*float64(ratio.Y)/float64(ratio.X))))
	var dst = image.NewRGBA(size)
	backgroundIndex += 1
	var backgroundColor = backgroundColors[backgroundIndex%len(backgroundColors)]
	draw.Draw(dst, dst.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)

	font, err := freetype.ParseFont(font_ttf())
	if err != nil {
		log.Println(err)
		return
	}

	var txtImage = image.NewRGBA(image.Rect(0, 0, 600, 600))
	draw.Draw(txtImage, txtImage.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)

	txtColor := image.White

	var fontSize = float64(imageReq.width) * 52 / 300 // Stupid magic number

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(fontSize)
	c.SetClip(txtImage.Bounds())
	c.SetDst(txtImage)
	c.SetSrc(txtColor)

	var offsetFix = int(math.Floor(fontSize * 12 / 72)) // Stupid magic number

	pt := freetype.Pt(0, int(c.PointToFix32(fontSize)>>8)-offsetFix)
	pt, err = c.DrawString(imageReq.ratio, pt)
	if err != nil {
		log.Println(err)
		return
	}

	txtSize := image.Pt(int(pt.X>>8), int(c.PointToFix32(fontSize)>>8)+2)

	txtBounds := image.Rect(
		int(math.Floor(float64(size.Max.X)/2.0)-(float64(txtSize.X)/2.0)),
		int(math.Floor(float64(size.Max.Y)/2.0)-(float64(txtSize.Y)/2.0)),
		int(math.Floor(float64(size.Max.X)/2.0)+(float64(txtSize.X)/2.0)),
		int(math.Floor(float64(size.Max.Y)/2.0)+(float64(txtSize.Y)/2.0)),
	)

	draw.Draw(dst, txtBounds, txtImage, image.ZP, draw.Src)

	if imageReq.format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{90})
		return
	}
	if imageReq.format == "png" {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
		return
	}

}

type ImageRequest struct {
	id     string
	ratio  string
	width  int
	format string
	dir    string
	path   string
}

func cp(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func minify(src, dst string) {
	cmd := exec.Command("imgmin", src, dst)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		log.Println(out.String())

		err = cp(src, dst)
		if err == nil {
			log.Println(err)
			return
		}
	}
	os.Remove(src)
}

func crop(w http.ResponseWriter, r *http.Request) {

	match, err := filepath.Match("/*/*/*.???", r.URL.Path)
	if !match {
		http.Error(w, "Couldn't find that.", 404)
		return
	}

	width, _ := strconv.Atoi(strings.Split(filepath.Base(r.URL.Path), ".")[0])
	imageReq := ImageRequest{
		width:  width,
		format: strings.Split(filepath.Base(r.URL.Path), ".")[1],
		ratio:  strings.Split(filepath.Dir(r.URL.Path), "/")[2],
		id:     strings.Split(filepath.Dir(r.URL.Path), "/")[1],
		dir:    filepath.Join(imageRoot, filepath.Dir(r.URL.Path)),
		path:   r.URL.Path,
	}

	if imageReq.width > 9000 {
		http.Error(w, "Enhance your calm, bro.", 420)
		return
	}

	_, err = os.Stat(filepath.Join(imageRoot, imageReq.id, "src"))
	if err != nil {
		if debug {
			placeholder(w, imageReq)
			return
		} else {
			http.Error(w, "Couldn't find that.", 404)
			return
		}
	}

	var dst = imageCrop(imageReq.id, imageReq.ratio)
	dst = imaging.Resize(dst, imageReq.width, 0, imaging.CatmullRom)

	err = os.MkdirAll(imageReq.dir, 0755)
	if err != nil {
		log.Print(imageReq.dir)
		log.Println(err)
	}

	croppedPath := filepath.Join(imageRoot, imageReq.path)
	if imgmin {
		croppedPath = croppedPath + ".preopt"
	}

	outputWriter, _ := os.Create(croppedPath)

	if imageReq.format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{75})
		if imgmin {
			jpeg.Encode(outputWriter, dst, &jpeg.Options{100})
		} else {
			jpeg.Encode(outputWriter, dst, &jpeg.Options{75})
		}
	}
	if imageReq.format == "png" {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
		png.Encode(outputWriter, dst)
	}

	if imgmin {
		go minify(croppedPath, filepath.Join(imageRoot, imageReq.path))
	}
}

func main() {
	flag.Parse()

	loadConfig()
	go buildIndex()

	var adminServeMux = http.NewServeMux()
	var adminServer = http.Server{
		Addr:    adminListen,
		Handler: adminServeMux,
	}

	var publicServeMux = http.NewServeMux()
	var publicServer = http.Server{
		Addr:    publicListen,
		Handler: publicServeMux,
	}

	publicServeMux.HandleFunc("/", crop)
	go func() {
		if debug {
			log.Print("Ready to crop! (debug enabled)")
		} else {
			log.Print("Ready to crop!")
		}
		publicServer.ListenAndServe()
	}()

	adminServeMux.HandleFunc("/cropper/js/", js)
	adminServeMux.HandleFunc("/cropper/css/", css)
	adminServeMux.HandleFunc("/cropper/", cropper)
	adminServeMux.HandleFunc("/api/new", new)
	adminServeMux.HandleFunc("/api/search", search)
	adminServeMux.HandleFunc("/api/", api)
	adminServer.ListenAndServe()
	adminReady = true
}
