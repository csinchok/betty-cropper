package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/pmylund/go-cache"
)

var BETTY_VERSION = "1.1.15"

var (
	version       = flag.Bool("version", false, "Print the version number and exit")
	configPath    = flag.String("config", "config.json", "Path for the config file")
	imageRoot     = "/var/betty-cropper"
	listen        = ":8888"
	publicAddress = "localhost:8888"
	debug         = false
	imgmin        = false
	ratios        []image.Point
	nextId        = -1
	adminReady    = false
	c             = cache.New(15*time.Minute, 30*time.Second)
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
		imgmin = false
	}

	type Config struct {
		ImageRoot     string   `json:"imageRoot"`     // Where we put out images
		Listen        string   `json:"listen"`        // The address that Betty listens on
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
	listen = config.Listen
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
	if r.Method != "GET" {
		http.Error(w, "GET only, you asshole.", 405)
		return
	}

	imageReq, err := ParseBettyRequest(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if imageReq.Width > 3000 {
		http.Error(w, "Couldn't find that.", 404)
		return
	}

	if imageReq.Width < 1 {
		http.Error(w, "Couldn't find that.", 404)
		return
	}

	img, err := imageReq.Image()
	if err != nil {
		if debug {
			placeholder(w, imageReq)
			return
		} else {
			http.Error(w, "Couldn't find that.", 404)
			return
		}
	}
	var selection = img.Selection(imageReq.RatioString)

	src, err := imaging.Open(filepath.Join(imageRoot, imageReq.Id, "src"))
	if err != nil {
		http.Error(w, "Couldn't find that.", 404)
		return
	}
	var dst = imaging.Crop(src, selection)
	dst = imaging.Resize(dst, imageReq.Width, 0, imaging.CatmullRom)

	err = os.MkdirAll(filepath.Dir(imageReq.Path()), 0755)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	croppedPath := imageReq.Path()
	if imgmin {
		croppedPath = croppedPath + ".preopt"
	}

	outputWriter, err := os.Create(croppedPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if imageReq.Format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, dst, &jpeg.Options{75})
		if imgmin {
			jpeg.Encode(outputWriter, dst, &jpeg.Options{100})
		} else {
			jpeg.Encode(outputWriter, dst, &jpeg.Options{80})
		}
	}
	if imageReq.Format == "png" {
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, dst)
		png.Encode(outputWriter, dst)
	}

	if imgmin {
		go minify(croppedPath, filepath.Join(imageRoot, imageReq.Path()))
	}
}

func main() {
	flag.Parse()

	loadConfig()
	go buildIndex()

	http.HandleFunc("/cropper/js/", js)
	http.HandleFunc("/cropper/css/", css)
	http.HandleFunc("/cropper/", cropper)
	http.HandleFunc("/api/new", new)
	http.HandleFunc("/api/search", search)
	http.HandleFunc("/api/", api)
	http.HandleFunc("/", crop)
	http.ListenAndServe(listen, nil)
	adminReady = true
}
