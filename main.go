package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pmylund/go-cache"
    "github.com/rafikk/imagick/imagick"
)

var BETTY_VERSION = "1.3.0"

type Config struct {
    ImageRoot          string   `json:"imageRoot"`     // Where we put out images
    Listen             string   `json:"listen"`        // The address that Betty listens on
    PublicAddress      string   `json:"publicAddress"` // The address that the public interface is served from
    Ratios             []string `json:"ratios"`        // A list of image ratios that we'll be cropping for
    PlaceholderEnabled bool     `json:"placeholderEnabled"`         // If debug is true or false
    PlaceholderFont    string   `json:"placeholderFont"`     // The font used for placeholder images
    CreditFont         string   `json:"creditFont"`    // The font used fo rhte image credits
    ElasticSearch      string   `json:"elasticsearch"` // An ElasticSearch URL
    ImgminEnabled      bool     `json:"imgminEnabled`  // imgmin emabled
}

var (
	version        = flag.Bool("version", false, "Print the version number and exit")
	configPath     = flag.String("config", "config.json", "Path for the config file")
	ratios         []image.Point
	nextId         = 1
	c              = cache.New(15*time.Minute, 30*time.Second)
    config         = Config{}
    redirectRegexp = regexp.MustCompile("^/([0-9]{5,})/((?:[0-9]+x[0-9]+)|original)/([0-9]+).(jpg|png)$")
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

	configBytes, err := ioutil.ReadFile(absConfigPath)
	if err != nil {
		log.Printf("Can't read the config file, because \"%s\", exiting.\n", err)
		os.Exit(1)
	}
	json.Unmarshal(configBytes, &config)

    if config.Listen == "" {
        config.Listen = ":8698"
    }

    if config.PublicAddress == "" {
        config.Listen = "http://localhost:8698"
    }

    if config.ImageRoot == "" {
        config.ImageRoot = "/var/betty-cropper"
    }

    // TODO: Could this even ever be nil?
    config.PlaceholderEnabled = true

    if config.PlaceholderFont == "" {
        config.PlaceholderFont = "Courier-New"
    }

    if config.CreditFont == "" {
        config.CreditFont = "Courier-New"
    }

	ratios = make([]image.Point, len(config.Ratios))
	// ratios = [config.Ratios.len()]image.Point
	for index, ratio := range config.Ratios {
		var w, _ = strconv.Atoi(strings.Split(ratio, "x")[0])
		var h, _ = strconv.Atoi(strings.Split(ratio, "x")[1])
		ratios[index] = image.Pt(w, h)
	}
}

func crop(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "GET only, you asshole.", 405)
		return
	}

	imageReq, err := ParseBettyRequest(r.URL.Path)
	if err != nil {
        re := *redirectRegexp
        var submatches = re.FindStringSubmatch(r.URL.Path)
        if submatches != nil {
            var location = fmt.Sprintf("%s/%s/%s/%s.%s", config.PublicAddress, GetRelImageDir(submatches[1]), submatches[2], submatches[3], submatches[4])
            http.Redirect(w, r, location, 301)
            return
        }
		http.Error(w, err.Error(), 500)
		return
	}

	if imageReq.Width > 3000 {
		http.Error(w, "Image too large", 420)
		return
	}

	if imageReq.Width < 1 {
		http.Error(w, "Image too small", 403)
		return
	}

	img, err := imageReq.Image()
	if err != nil {
		if config.PlaceholderEnabled {
			placeholder(w, imageReq)
			return
		} else {
			http.Error(w, "Couldn't find that.", 404)
			return
		}
	}
	var selection = img.Selection(imageReq.RatioString)

    imagick.Initialize()
    defer imagick.Terminate()
    mw := imagick.NewMagickWand()
    defer mw.Destroy()

    // Read the image
    err = mw.ReadImage(filepath.Join(GetImageDir(img.Id), "src"))
    if err != nil {
        http.Error(w, "Couldn't find that.", 404)
    }

    // Crop to selection
    width := uint(selection.Max.X - selection.Min.X)
    height := uint(selection.Max.Y - selection.Min.Y)
    mw.CropImage(width, height, selection.Min.X, selection.Min.Y)
    err = mw.ResizeImage(uint(imageReq.Width), uint(imageReq.Height()), imagick.FILTER_LANCZOS, 1)
    if err != nil {
        log.Println(err)
    }

    if img.Credit != ""  && imageReq.Width > 250 {
        // TODO: Also check image width?
        dw := imagick.NewDrawingWand()

        dw.SetFont(config.PlaceholderFont)
        dw.SetFontSize(12)

        pw := imagick.NewPixelWand()
        pw.SetColor("#FFFFFF")
        dw.SetFillColor(pw)
        pw.Destroy()

        fontMetrics := mw.QueryFontMetrics(dw, img.Credit)

        scale := float64(imageReq.Width) / float64(width)
        // log.Printf("DY: %d", imageReq.Height() + selection.Min.Y - 10)
        dy := (float64(selection.Max.Y) * scale) - 10
        dx := (float64(selection.Max.X) * scale) - fontMetrics.TextWidth - 10
        err = mw.AnnotateImage(dw, dx, dy, 0, img.Credit)
        if err != nil {
            log.Println(err)
        }
        err = mw.DrawImage(dw)
        if err != nil {
            log.Println(err)
        }
        dw.Destroy()
    }

	err = os.MkdirAll(filepath.Dir(imageReq.Path()), 0755)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	outputFile, err := os.Create(imageReq.Path())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if imageReq.Format == "jpg" {
        mw.SetImageFormat("jpeg")
        mw.SetImageCompressionQuality(uint(img.MinQuality))
		w.Header().Set("Content-Type", "image/jpeg")
	}
	if imageReq.Format == "png" {
        mw.SetImageFormat("png")
		w.Header().Set("Content-Type", "image/png")
	}
    imageBytes := mw.GetImageBlob()
    w.Write(imageBytes)
    outputFile.Write(imageBytes)

    return
}

func main() {
	flag.Parse()

	loadConfig()
	go buildIndex()

	http.HandleFunc("/api/new", new)
	http.HandleFunc("/api/search", search)
	http.HandleFunc("/api/", api)
	http.HandleFunc("/", crop)
	http.ListenAndServe(config.Listen, nil)
}
