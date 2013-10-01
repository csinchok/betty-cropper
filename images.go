package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"
    "regexp"
	"path/filepath"
	"strconv"
	"strings"
    "log"

    "github.com/rafikk/imagick/imagick"
    "github.com/csinchok/imgmin-go"
)

// The regex that validates and parses incoming image requests.
var imageRegexp = regexp.MustCompile("^(?P<image_id_path>(?:/[0-9]{1,4})+)/(?P<ratio>(?:[0-9]+x[0-9]+)|original)/(?P<width>[0-9]+).(?P<format>jpg|png)$")


type BettyImage struct {
	Id         string
	Credit     string
	Filename   string
	Selections map[string]image.Rectangle
	Size       image.Point
    MinQuality int
}

// Because of filesystem limitations, if we have more than 9999
// images, we need to split things into subdirectories. This function
// gives us the base directory for a given image id.
func GetImageDir(imageId string) string {
	return filepath.Join(config.ImageRoot, GetRelImageDir(imageId))
}

func cleanImageName(s string) string {
    return strings.Replace(s, " ", "_", -1)
}

func expandImageName(s string) string {
    return strings.Replace(s, "_", " ", -1)
}

// Just used for redirects....
func GetRelImageDir(imageId string) string {
    var buffer bytes.Buffer
    for index, value := range imageId {
        buffer.WriteRune(value)
        if (index+1)%4 == 0 {
            buffer.WriteString("/")
        }
    }
    return buffer.String()
}

// This function retrieves and caches the info for an image id.
func GetBettyImage(imageId string) (BettyImage, error) {
	data, found := c.Get(imageId)
	if found {
		return data.(BettyImage), nil
	}

	imageDir := GetImageDir(imageId)
	srcPath := filepath.Join(imageDir, "src")
	dstPath, err := os.Readlink(srcPath)
	if err != nil {
		// If this fails, this iamge doesn't exist, so we bail.
		return BettyImage{}, err
	}

	// If we got here, this is *likely* a real image, so we'll start filling it out.
	img := BettyImage{
		Id:       imageId,
		Filename: filepath.Base(dstPath),
	}

    imagick.Initialize()
    mw := imagick.NewMagickWand()

    // Load up the original, get the size.
    err = mw.ReadImage(filepath.Join(imageDir, "src"))
    if err != nil {
        return BettyImage{}, err
    }
    img.Size = image.Pt(int(mw.GetImageWidth()), int(mw.GetImageHeight()))
    mw.Destroy()
    imagick.Terminate()

	// Look for a credit.txt, store that info if it exists.
	creditPath := filepath.Join(imageDir, "credit.txt")
	creditBytes, err := ioutil.ReadFile(creditPath)
	if err == nil {
		img.Credit = string(creditBytes)
	}

    qualityPath := filepath.Join(imageDir, "quality.txt")
    qualityBytes, err := ioutil.ReadFile(qualityPath)
    if err == nil {
        img.MinQuality, _ = strconv.Atoi(string(qualityBytes))
    } else {
        img.MinQuality = 75
        go findOptimalQuality(img.Id)
    }

	// Look for the selections.json file, store that info if it exists.
	var selectionsJsonPath = filepath.Join(imageDir, "selections.json")
	// var selections map[string]image.Rectangle
	selectionBytes, err := ioutil.ReadFile(selectionsJsonPath)
	if err == nil {
		json.Unmarshal(selectionBytes, &img.Selections)
	} else {
		img.Selections = make(map[string]image.Rectangle, len(ratios))
	}

	// Put this image into the cache, and return
	c.Set(imageId, img, 0)
	return img, nil
}

// Get a human readable name for the image, taken from the filename.
func (img BettyImage) Name() string {
	imageName := strings.Replace(img.Filename, filepath.Ext(img.Filename), "", 1)
	return strings.Replace(imageName, "_", " ", -1)
}

func (img BettyImage) Dir() string {
    return GetImageDir(img.Id)
}

func findOptimalQuality(imageId string) {

    imagick.Initialize()
    mw := imagick.NewMagickWand()

    // Read the image
    err := mw.ReadImage(filepath.Join(GetImageDir(imageId), "src"))
    if err != nil {
        log.Println(err.Error())
        return
    }

    if mw.GetImageFormat() != "jpeg" {
        mw.SetImageFormat("jpeg")
        imageBytes := mw.GetImageBlob()
        mw.ReadImageBlob(imageBytes)
    } 

    out, err := imgmin.SearchQuality(mw, imgmin.Options{})
    if err != nil {
        log.Println(err.Error())
        return
    }
    qualityString := strconv.Itoa(int(out.GetImageCompressionQuality()))
    
    img, err := GetBettyImage(imageId)
    img.MinQuality = int(out.GetImageCompressionQuality())
    c.Set(img.Id, img, 0)

    qualityPath := filepath.Join(GetImageDir(imageId), "quality.txt")
    err = ioutil.WriteFile(qualityPath, []byte(qualityString), 0644)
    if err != nil {
        log.Println(err.Error())
        return
    }
}

func clearCrop(imageId string, ratio string) {
    ratioDir := filepath.Join(GetImageDir(imageId), ratio)
    f, err := os.Open(ratioDir)
    if err != nil {
        return
    }
    list, err := f.Readdir(-1)
    f.Close()
    if err != nil {
        return
    }
    for _,crop := range list {
        os.Remove(filepath.Join(ratioDir, crop.Name()))
    }
}

// Given a ratio string, get the selection that we'll be cropping to, either
// from the selections.json file, or just from the middle of the iamge.
func (img BettyImage) Selection(ratioString string) image.Rectangle {
	// If this selection is specified, just return it.
	if selection, ok := img.Selections[ratioString]; ok {
		return selection
	}

	// The selection for this ratio hasn't been set. Let's just use
	// the middle of the image.
	var ratio = img.Size
	if ratioString != "original" {
		var w, _ = strconv.Atoi(strings.Split(ratioString, "x")[0])
		var h, _ = strconv.Atoi(strings.Split(ratioString, "x")[1])
		ratio = image.Point{w, h}
	}

	var originalRatio = float64(img.Size.X) / float64(img.Size.Y)
	var selectionRatio = float64(ratio.X) / float64(ratio.Y)

	var min = image.Pt(0, 0)
	var max = img.Size

	if selectionRatio < originalRatio {
		var xOffset = (float64(img.Size.X) - (float64(img.Size.Y) * float64(ratio.X) / float64(ratio.Y))) / 2.0
		min = image.Pt(int(math.Floor(xOffset)), 0)
		max = image.Pt(img.Size.X-int(math.Floor(xOffset)), img.Size.Y)
	}
	if selectionRatio > originalRatio {
		var yOffset = (float64(img.Size.Y) - (float64(img.Size.X) * float64(ratio.Y) / float64(ratio.X))) / 2.0

		min = image.Pt(0, int(math.Floor(yOffset)))
		max = image.Pt(img.Size.X, img.Size.Y-int(math.Floor(yOffset)))
	}

	return image.Rectangle{min, max}
}

// Update a selection for an image, caching the selections, and also writing them to disk.
func (img *BettyImage) SetSelection(ratioString string, selection image.Rectangle) error {

    // Update the selection
    img.Selections[ratioString] = selection

    // Serialize it, and write it out to disk
    data, err := json.Marshal(img.Selections)
    if err != nil {
        return err
    }
    selectionsJsonPath := filepath.Join(GetImageDir(img.Id), "selections.json")
    err = ioutil.WriteFile(selectionsJsonPath, data, 0644)
    if err != nil {
        return err
    }

    // Cache it
    c.Set(img.Id, *img, 0)

    go clearCrop(img.Id, ratioString)

    return nil
}

func (img *BettyImage) SetName(filename string) error {

    // Delete the old link, add a new one
    srcPath := filepath.Join(GetImageDir(img.Id), "src")
    dst, err := os.Readlink(srcPath)
    if err != nil {
        return err
    }

    oldPath := filepath.Join(GetImageDir(img.Id), dst)  // The original is linked using a relative path.
    newName := filename + filepath.Ext(oldPath)  // The new name should have the same extension as the old one
    newPath := filepath.Join(GetImageDir(img.Id), cleanImageName(newName))  // Let's make sure we don't have weird chars
    
    err = os.Rename(oldPath, newPath)  // Move the original
    if err != nil {
        return err
    }

    os.Remove(srcPath)  // Remove the src link
    if err != nil {
        return err
    }

    err = os.Symlink(newPath, srcPath)  // Add the new link
    if err != nil {
        return err
    }

    img.Filename = newName  // Update the object
    c.Set(img.Id, *img, 0) // Cache it
    return nil
}

func (img *BettyImage) SetCredit(credit string) error {
    creditPath := filepath.Join(GetImageDir(img.Id), "credit.txt")
    creditFile, err := os.Create(creditPath)
    if err != nil {
        return err
    }
    err = creditFile.Truncate(0)

    _, err = creditFile.WriteString(credit)
    if err != nil {
        return err
    }

    img.Credit = credit  // Update the object
    c.Set(img.Id, *img, 0) // Cache it
    return nil
}
type ImageSize struct {
    Width    int  `json:"width"`
    Height   int  `json:"height"`
}
type SearchResult struct {
    Id         string                     `json:"id"`
    Name       string                     `json:"name"`
    Filename   string                     `json:"filename"`
    Credit     string                     `json:"credit,omitempty"`
    Size       ImageSize                  `json:"size"`
    Selections map[string]image.Rectangle `json:"selections"`
}

func (img BettyImage) Serialized() SearchResult {
    result := SearchResult{
        Id: img.Id,
        Credit: img.Credit,
        Filename: img.Filename,
        Name: img.Name(),
        Size: ImageSize{Width: img.Size.X, Height: img.Size.Y},
        Selections: make(map[string]image.Rectangle),
    }
    for _, ratio := range ratios {
        ratioString := fmt.Sprintf("%dx%d", ratio.X, ratio.Y)
        result.Selections[ratioString] = img.Selection(ratioString)
    }
    return result
}

// The Betty request struct holds information about a crop
// request, and is created from a URL.Path
type BettyRequest struct {
	Id          string
	RatioString string
	Width       int
	Format      string
}

// Get the BettyImage associated with this request.
func (r BettyRequest) Image() (BettyImage, error) {
	return GetBettyImage(r.Id)
}

// Get the absolute path to the file we're going ot create.
func (r BettyRequest) Path() string {
	var filename = fmt.Sprintf("%d.%s", r.Width, r.Format)
	return filepath.Join(GetImageDir(r.Id), r.RatioString, filename)
}

func (r BettyRequest) Height() int {
    return int(math.Floor(float64(r.Width) * float64(r.Ratio().Y) / float64(r.Ratio().X)))
}

// Get the size of the output image
func (r BettyRequest) Size() image.Rectangle {
	return image.Rect(0, 0, r.Width, r.Height())
}

// Represent the image ratio as an image.Point.
func (r BettyRequest) Ratio() image.Point {
	if r.RatioString == "original" {
        img, err := r.Image()
        if err == nil {
            return img.Size
        } else {
            return image.Point{}
        }
		
	}
	var w, _ = strconv.Atoi(strings.Split(r.RatioString, "x")[0])
	var h, _ = strconv.Atoi(strings.Split(r.RatioString, "x")[1])
	return image.Point{w, h}
}

// Parse a URL.Path into BettyRequest, checking to make sure the URL is valid.
func ParseBettyRequest(URLPath string) (BettyRequest, error) {
	re := *imageRegexp
	var submatches = re.FindStringSubmatch(URLPath)
	if submatches == nil {
		return BettyRequest{}, errors.New("Bad image request")
	}
	width, err := strconv.Atoi(submatches[3])
	if err != nil {
		return BettyRequest{}, err
	}
	var imageReq = BettyRequest{
		Id:          strings.Join(strings.Split(submatches[1], "/"), ""),
		RatioString: submatches[2],
		Width:       width,
		Format:      submatches[4],
	}
	return imageReq, nil
}
