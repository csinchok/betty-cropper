package main

import (
	"bytes"
	"encoding/json"
    "errors"
	"flag"
	"fmt"
	"image"
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
    "regexp"
	"strconv"
	"strings"

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
	listen   = ":8888"
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
		imgmin = false
	}

	type Config struct {
		ImageRoot    string `json:"imageRoot"`    // Where we put out images
		Listen  string `json:"listen"`  // The address that Betty listens on
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

type ImageRequest struct {
    Id     string
    RatioString  string
    Width  int
    Format string
}

func (r ImageRequest) Dir() string {
    var buffer bytes.Buffer
    for index, value := range r.Id {
        buffer.WriteRune(value)
        if (index + 1) % 4 == 0 {
            buffer.WriteString("/")
        }
    }
    return filepath.Join(imageRoot, buffer.String());
}

func (r ImageRequest) Path() string {
    var filename = fmt.Sprintf("%d.%s", r.Width, r.Format)
    return filepath.Join(r.Dir(), r.RatioString, filename)
}

func (r ImageRequest) Size() image.Rectangle {
    var height = int( math.Floor( float64(r.Width) * float64(r.Ratio().Y) / float64(r.Ratio().X) ) )
    return image.Rect(0, 0, r.Width, height)
}

func (r ImageRequest) Ratio() image.Point {
    if r.RatioString == "original" {
        return image.Point{}
    }
    var w, _ = strconv.Atoi(strings.Split(r.RatioString, "x")[0])
    var h, _ = strconv.Atoi(strings.Split(r.RatioString, "x")[1])
    return image.Point{w, h}
}

func (r ImageRequest) Selection(imageSize image.Point) image.Rectangle {
    var selectionJsonPath = filepath.Join(imageRoot, r.Id, "selections.json")

    var selections map[string]image.Rectangle
    selectionBytes, err := ioutil.ReadFile(selectionJsonPath)
    if err == nil {
        json.Unmarshal(selectionBytes, &selections)
    } else {
        selections = make(map[string]image.Rectangle, len(ratios))
    }
    // TODO: maybe pss the string into this?
    if selection, ok := selections[r.RatioString]; ok {
        return selection
    }

    var ratio = imageSize
    if r.RatioString != "original" {
        ratio = r.Ratio()
    }

    var originalRatio = float64(imageSize.X) / float64(imageSize.Y)
    var selectionRatio = float64(ratio.X) / float64(ratio.Y)

    var min = image.Pt(0, 0)
    var max = imageSize

    if selectionRatio < originalRatio {
        var xOffset = (float64(imageSize.X) - (float64(imageSize.Y) * float64(ratio.X) / float64(ratio.Y))) / 2.0
        min = image.Pt(int(math.Floor(xOffset)), 0)
        max = image.Pt(imageSize.X - int(math.Floor(xOffset)), imageSize.Y)
    }
    if selectionRatio > originalRatio {
        var yOffset = (float64(imageSize.Y) - (float64(imageSize.X) * float64(ratio.Y) / float64(ratio.X))) / 2.0

        min = image.Pt(0, int(math.Floor(yOffset)))
        max = image.Pt(imageSize.X, imageSize.Y - int(math.Floor(yOffset)))
    }

    return image.Rectangle{min, max}

}

var imageRegexp = regexp.MustCompile("^(?P<image_id_path>(?:/[0-9]{1,4})+)/(?P<ratio>(?:[0-9]+x[0-9]+)|original)/(?P<width>[0-9]+).(?P<format>jpg|png)$")

func NewImageRequest(URLPath string) (ImageRequest, error) {
    re := *imageRegexp
    var submatches = re.FindStringSubmatch(URLPath)
    if submatches == nil {
        return ImageRequest{}, errors.New("Bad image request")
    }
    width, err := strconv.Atoi(submatches[3])
    if err != nil {
        return ImageRequest{}, err
    }
    var imageReq = ImageRequest{
        Id: strings.Join(strings.Split(submatches[1], "/"), ""),
        RatioString: submatches[2],
        Width: width,
        Format: submatches[4],
    }
    return imageReq, nil
}

func crop(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "GET only, you asshole.", 405)
        return
    }

    imageReq, err := NewImageRequest(r.URL.Path)
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

	_, err = os.Stat(filepath.Join(imageRoot, imageReq.Id, "src"))
	if err != nil {
		if debug {
			placeholder(w, imageReq)
			return
		} else {
			http.Error(w, "Couldn't find that.", 404)
			return
		}
	}

    src, err := imaging.Open(filepath.Join(imageRoot, imageReq.Id, "src"))
    if err != nil {
        fmt.Println("Couldn't find an image. Did you set the image root?")
    }
    var selection = imageReq.Selection(src.Bounds().Max)

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
