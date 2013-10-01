package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"
    "errors"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/argusdusty/Ferret"
    "github.com/rafikk/imagick/imagick"
    // "github.com/csinchok/imgmin-go"
)

var SearchEngine *ferret.InvertedSuffix

type IndexedImage struct {
    Id      string
	Name    string
}

func searchQuality(imageId string) {
    
}

func buildIndex() {
    var ids = make([]string, 0)
    var names = make([]string, 0)
    var datums = make([]interface{}, 0)
    var count = 0
    filepath.Walk(config.ImageRoot, func(path string, info os.FileInfo, err error) error {
        if count > 10000 {
            return errors.New("Ferret can't handle that many images, please use ES.")
        }

        if filepath.Base(path) == "src" {
            dir, err := filepath.Rel(config.ImageRoot, filepath.Dir(path))
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

            ids = append(ids, data.Id)
            names = append(names, data.Name)
            datums = append(datums, data)
            count += 1
            if count % 1000 == 0 {
                log.Printf("Crawled %d items...", count)
            }


            id, err := strconv.Atoi(data.Id)
            if err == nil && id >= nextId {
                nextId = id + 1
            }
        }
        return nil
    })
    log.Println("Indexing finished!")
    SearchEngine = ferret.New(names, ids, datums, ferret.UnicodeToLowerASCII)
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

    // Initialize a wand
    imagick.Initialize()
    defer imagick.Terminate()
    mw := imagick.NewMagickWand()
    defer mw.Destroy()

    // Make sure this is a valid image.
    err = mw.ReadImageBlob(imgBytes)
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
        Size: image.Pt(int(mw.GetImageWidth()), int(mw.GetImageHeight())),
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
	err = os.Symlink(img.Filename, srcLinkPath)
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