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

type SearchResult struct {
	ImageId string `json:"imageId"`
	Name    string `json:"name"`
	Credit  string `json:"credit"`
}

func cleanImageName(s string) string {
	return strings.Replace(s, " ", "_", -1)
}

func expandImageName(s string) string {
	return strings.Replace(s, "_", " ", -1)
}

func buildIndex() {
	SearchEngine = ferret.New(make([]string, 0), make([]string, 0), make([]interface{}, 0), ferret.UnicodeToLowerASCII)

	dirList, err := ioutil.ReadDir(imageRoot)
	if err != nil {
		log.Fatal(err)
	}
	if len(dirList) == 0 {
		nextId = 1
	} else {
		for _, dir := range dirList {
			srcPath := filepath.Join(imageRoot, dir.Name(), "src")

			var creditString string
			creditPath := filepath.Join(imageRoot, dir.Name(), "credit.txt")
			creditBytes, err := ioutil.ReadFile(creditPath)
			if err == nil {
				creditString = string(creditBytes)
			}

			dest, err := os.Readlink(srcPath)
			if err == nil {
				imageId, err := strconv.Atoi(dir.Name())
				if err == nil {
					if imageId >= nextId {
						nextId = imageId + 1
					}
				} else {
					log.Println(err.Error())
				}
				filename := filepath.Base(dest)
				imageName := strings.Replace(filename, filepath.Ext(filename), "", 1)
				data := SearchResult{
					Name:    expandImageName(imageName),
					ImageId: dir.Name(),
					Credit:  creditString,
				}
				SearchEngine.Insert(filepath.Base(dest), dir.Name(), data)
			}
		}
	}
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

	ids, values := SearchEngine.Query(query, 25)
	var results []SearchResult = make([]SearchResult, len(ids))
	for index, id := range ids {
		data := values[index]
		results[index] = SearchResult{
			ImageId: id,
			Name:    data.(SearchResult).Name,
			Credit:  data.(SearchResult).Credit,
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

		log.Println(r.URL.Path)

		if r.Method != "POST" {
			http.Error(w, "POST only, you asshole.", 405)
			return
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			fmt.Fprintln(w, 200)
			return
		}

		var imageId = filepath.Base(filepath.Dir(r.URL.Path))
		var imageRatio = filepath.Base(r.URL.Path)

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
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}

	if matched, _ := filepath.Match("/api/*", r.URL.Path); matched {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		var imageId = filepath.Base(r.URL.Path)
		_ = imageId

		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			fmt.Fprintln(w, "")
		}

		if r.Method == "GET" {
			srcPath := filepath.Join(imageRoot, imageId, "src")
			originalPath, err := os.Readlink(srcPath)
			filename := filepath.Base(originalPath)
			name := strings.Replace(filename, filepath.Ext(filename), "", 1)

			var creditString string
			creditPath := filepath.Join(imageRoot, imageId, "credit.txt")
			creditBytes, err := ioutil.ReadFile(creditPath)
			if err == nil {
				creditString = string(creditBytes)
			}

			imageData := SearchResult{
				ImageId: imageId,
				Name:    expandImageName(name),
				Credit:  creditString,
			}
			data, err := json.Marshal(imageData)
			w.WriteHeader(200)
			w.Write(data)
			return
		}

		if r.Method == "POST" {
			if r.FormValue("name") != "" {
				srcPath := filepath.Join(imageRoot, imageId, "src")
				oldPath, err := os.Readlink(srcPath)
				if err == nil {
					newName := r.FormValue("name") + filepath.Ext(oldPath)
					newPath := filepath.Join(imageRoot, imageId, cleanImageName(newName))
					os.Rename(oldPath, newPath)
					os.Remove(srcPath)
					err = os.Symlink(newPath, srcPath)
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
				} else {
					log.Println(err)
				}
			}

			if creditString := r.FormValue("credit"); creditString != "" {
				creditPath := filepath.Join(imageRoot, imageId, "credit.txt")
				creditFile, err := os.Create(creditPath)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				err = creditFile.Truncate(0)

				_, err = creditFile.WriteString(creditString)
				if err != nil {
					log.Println(err)
				}
			}

			w.WriteHeader(200)
			fmt.Fprintln(w, "")

			// This sucks, but we need to totally rebuild the index after we update anything
			go buildIndex()
			return
		}

		return
	}
}

func new(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With")

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

	var filename = fileHeader.Filename
	if r.FormValue("name") != "" {
		filename = r.FormValue("name") + filepath.Ext(filename)
	}

	srcData, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "File error", 500)
		return
	}

	var imageId = strconv.Itoa(nextId)
	nextId += 1

	err = os.MkdirAll(filepath.Join(imageRoot, imageId), 0755)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var srcPath = filepath.Join(imageRoot, imageId, cleanImageName(filename))
	var srcLinkPath = filepath.Join(imageRoot, imageId, "src")

	err = ioutil.WriteFile(srcPath, srcData, 0777)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = os.Symlink(srcPath, srcLinkPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := SearchResult{
		Name:    filename,
		ImageId: imageId,
		Credit:  "",
	}
	SearchEngine.Insert(filename, imageId, data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "{\"id\":\"%s\"}", imageId)
}

func cropper(w http.ResponseWriter, r *http.Request) {
	var imageId = strings.Split(r.URL.Path, "/")[2]
	img, err := GetBettyImage(imageId)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
