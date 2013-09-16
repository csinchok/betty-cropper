package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/argusdusty/Ferret"
	// "github.com/disintegration/imaging"
)

var SearchEngine *ferret.InvertedSuffix

type IndexedImage struct {
    Id      string
	Name    string
}

func buildIndex() {
	SearchEngine = ferret.New(make([]string, 0), make([]string, 0), make([]interface{}, 0), ferret.UnicodeToLowerASCII)
    var count = 1
    filepath.Walk(imageRoot, func(path string, info os.FileInfo, err error) error {
        if filepath.Base(path) == "src" {
            dir, err := filepath.Rel(imageRoot, filepath.Dir(path))
            if err != nil {
                return err
            }
            dstPath, err := os.Readlink(path)
            if err != nil {
                return err
            }
            data := IndexedImage{
                Id: strings.Join(strings.Split(dir, "/"), ""),
                Name: expandImageName(filepath.Base(dstPath)),
            }
            id, err := strconv.Atoi(data.Id)
            if err == nil && id >= nextId {
                nextId = id + 1
            }
            count += 1
            if count % 100 == 0 {
                log.Printf("Indexed %d items...\n", count)
            }
            SearchEngine.Insert(data.Name, data.Id, data)
        }
        return nil
    })
}

func search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		fmt.Fprintln(w, "")
		return
	}

	if r.Method != "GET" {
		http.Error(w, "GET only, you asshole.", 405)
		return
	}

	queryList, ok := r.URL.Query()["q"]
	var query = ""
	if ok {
		query = queryList[0]
	}

	ids, _ := SearchEngine.Query(query, 25)
	var results []SearchResult = make([]SearchResult, len(ids))
	for index, id := range ids {
        img, err := GetBettyImage(id)
        if err == nil {
            results[index] = img.Serialized()
        } else {
            results[index] = SearchResult{}
        }
	}

	b, err := json.Marshal(results)
	if err != nil {
		log.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(b)
	return
}

func api(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.WriteHeader(200)
		fmt.Fprintln(w, 200)
		return
	}

	if matched, _ := filepath.Match("/api/*/*", r.URL.Path); matched {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			fmt.Fprintln(w, 200)
			return
		}
        if r.Method != "POST" {
            http.Error(w, "POST only, you asshole.", 405)
            return
        }

		var imageId = filepath.Base(filepath.Dir(r.URL.Path))
		var imageRatio = filepath.Base(r.URL.Path)

        img, err := GetBettyImage(imageId)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

		minX, err := strconv.Atoi(r.FormValue("minX"))
		if err != nil {
            http.Error(w, err.Error(), 500)
            return
		}
		minY, err := strconv.Atoi(r.FormValue("minY"))
		if err != nil {
            http.Error(w, err.Error(), 500)
            return			
		}
		maxX, err := strconv.Atoi(r.FormValue("maxX"))
		if err != nil {
            http.Error(w, err.Error(), 500)
            return
		}
		maxY, err := strconv.Atoi(r.FormValue("maxY"))
		if err != nil {
			http.Error(w, err.Error(), 500)
            return
		}

		var selection = image.Rect(minX, minY, maxX, maxY)
        err = img.SetSelection(imageRatio, selection)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		return
	}

	if matched, _ := filepath.Match("/api/*", r.URL.Path); matched {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		var imageId = filepath.Base(r.URL.Path)
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			fmt.Fprintln(w, "")
		}

        img, err := GetBettyImage(imageId)
        if err != nil {
            http.Error(w, err.Error(), 404)
            return
        }

		if r.Method == "GET" {
			data, _ := json.Marshal(img.Serialized())
			w.WriteHeader(200)
			w.Write(data)
			return
		}

		if r.Method == "POST" {
			if r.FormValue("name") != "" {
                err = img.SetName(r.FormValue("name"))
                if err != nil {
                    http.Error(w, err.Error(), 500)
                    return
                }
			}

			if creditString := r.FormValue("credit"); creditString != "" {
                err = img.SetCredit(creditString)
                if err != nil {
                    http.Error(w, err.Error(), 500)
                    return
                }
			}

            data, _ := json.Marshal(img.Serialized())
            w.WriteHeader(200)
            w.Write(data)
            return

			// This sucks, but we need to totally rebuild the index after we update anything
			go buildIndex()
			return
		}

		http.Error(w, "Couldn't find that", 404)
        return
	}
}

func new(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		fmt.Fprintln(w, 200)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "POST only, you asshole.", 405)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	// TODO: check to make sure it's a valid image
	if err != nil {
		http.Error(w, "POST error", 500)
		return
	}

    imgBytes, err := ioutil.ReadAll(file)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // The previous read went to the end of the file, so let's go to the start again.
    _, err = file.Seek(0, 0)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Make sure this is a valid image.
    imgData, _, err := image.Decode(file)
    if err != nil {
        http.Error(w, "File error", 500)
        return
    }

	var filename = fileHeader.Filename
	if r.FormValue("name") != "" {
		filename = r.FormValue("name") + filepath.Ext(filename)
	}

    img := BettyImage{
        Id: strconv.Itoa(nextId),
        Filename: cleanImageName(filename),
        Size: imgData.Bounds().Max,
    }

    nextId += 1  // TODO: Add mutex here?

	err = os.MkdirAll(GetImageDir(img.Id), 0755)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var srcPath = filepath.Join(GetImageDir(img.Id), img.Filename)
	var srcLinkPath = filepath.Join(GetImageDir(img.Id), "src")

	err = ioutil.WriteFile(srcPath, imgBytes, 0644)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = os.Symlink(srcPath, srcLinkPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	indexedImage := IndexedImage{
		Name: img.Name(),
		Id:   img.Id,
	}
	SearchEngine.Insert(img.Name(), img.Id, indexedImage)

	w.Header().Set("Content-Type", "application/json")
    data, _ := json.Marshal(img.Serialized())
    w.WriteHeader(201)
    w.Write(data)
    return
}

func cropper(w http.ResponseWriter, r *http.Request) {
	var imageId = strings.Split(r.URL.Path, "/")[2]
	img, err := GetBettyImage(imageId)
	if err != nil {
		http.Error(w, "Couldn't find that", 404)
        return
	}

	var imageScale = 600.0 / float64(img.Size.X)

	var selections = make([]image.Rectangle, len(ratios))

	for i, ratio := range ratios {
		ratioString := fmt.Sprintf("%dx%d", ratio.X, ratio.Y)
		selections[i] = img.Selection(ratioString)
	}

	var scaledSize = image.Pt(600, int(600.0*float64(img.Size.Y)/float64(img.Size.X)))

	t := template.New("cropper.html")
	t.Parse(string(html_cropper_html()))
	t.Execute(w, map[string]interface{}{
		"ImageId":       imageId,
		"Ratios":        ratios,
		"Selections":    selections,
		"ScaledSize":    scaledSize,
		"ImageScale":    imageScale,
		"PublicAddress": publicAddress,
	})
}

func js(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/cropper/js/jquery.color.js" {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(js_jquery_color_js())
		return
	}
	if r.URL.Path == "/cropper/js/jquery.Jcrop.min.js" {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(js_jquery_jcrop_min_js())
		return
	}
	http.Error(w, "Couldn't find that.", 404)
}

func css(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/cropper/css/Jcrop.gif" {
		w.Header().Set("Content-Type", "image/gif")
		w.Write(css_jcrop_gif())
		return
	}
	if r.URL.Path == "/cropper/css/jquery.Jcrop.min.css" {
		w.Header().Set("Content-Type", "text/css")
		w.Write(css_jquery_jcrop_min_css())
		return
	}
	http.Error(w, "Couldn't find that.", 404)
}
