package main


import (
    "os"
    "image"
    "bytes"
    "path/filepath"
    "github.com/disintegration/imaging"
    "io/ioutil"
    "encoding/json"
    "strconv"
    "strings"
    "math"
    "fmt"
    "errors"
)


type BettyImage struct {
    Id         string
    Credit     string
    Filename   string
    Selections map[string]image.Rectangle
    Size       image.Point
}

// Because of filesystem limitations, if we have more than 9999
// images, we need to split things into subdirectories. This function
// gives us the base directory for a given image id.
func GetImageDir(imageId string) string {
    var buffer bytes.Buffer
    for index, value := range imageId {
        buffer.WriteRune(value)
        if (index + 1) % 4 == 0 {
            buffer.WriteString("/")
        }
    }
    return filepath.Join(imageRoot, buffer.String());
}

// This function retrieves and caches the image info.
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
        Id: imageId,
        Filename: filepath.Base(dstPath),
    }

    // Load up the original, get the size.
    src, err := imaging.Open(filepath.Join(imageDir, "src"))
    if err != nil {
        return BettyImage{}, err
    }
    img.Size = src.Bounds().Max

    // Look for a credit.txt, store that info if it exists.
    creditPath := filepath.Join(imageDir, "credit.txt")
    creditBytes, err := ioutil.ReadFile(creditPath)
    if err == nil {
        img.Credit = string(creditBytes)
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

func (img BettyImage) Open() (image.Image, error) {
    return imaging.Open(filepath.Join(GetImageDir(img.Id), "src"))
}

func (img BettyImage) Name() string {
    imageName := strings.Replace(img.Filename, filepath.Ext(img.Filename), "", 1)
    return strings.Replace(imageName, "_", " ", -1)
}

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
        max = image.Pt(img.Size.X - int(math.Floor(xOffset)), img.Size.Y)
    }
    if selectionRatio > originalRatio {
        var yOffset = (float64(img.Size.Y) - (float64(img.Size.X) * float64(ratio.Y) / float64(ratio.X))) / 2.0

        min = image.Pt(0, int(math.Floor(yOffset)))
        max = image.Pt(img.Size.X, img.Size.Y - int(math.Floor(yOffset)))
    }

    return image.Rectangle{min, max}
}

type BettyRequest struct {
    Id     string
    RatioString  string
    Width  int
    Format string
}

func (r BettyRequest) Image() (BettyImage, error) {
    return GetBettyImage(r.Id)
}

func (r BettyRequest) Path() string {
    var filename = fmt.Sprintf("%d.%s", r.Width, r.Format)
    return filepath.Join(GetImageDir(r.Id), r.RatioString, filename)
}

func (r BettyRequest) Size() image.Rectangle {
    var height = int( math.Floor( float64(r.Width) * float64(r.Ratio().Y) / float64(r.Ratio().X) ) )
    return image.Rect(0, 0, r.Width, height)
}

func (r BettyRequest) Ratio() image.Point {
    if r.RatioString == "original" {
        return image.Point{}
    }
    var w, _ = strconv.Atoi(strings.Split(r.RatioString, "x")[0])
    var h, _ = strconv.Atoi(strings.Split(r.RatioString, "x")[1])
    return image.Point{w, h}
}


// This isn't really a "New" request, it's just parsing a URL. Do we even need it?
func NewBettyRequest(URLPath string) (BettyRequest, error) {
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
        Id: strings.Join(strings.Split(submatches[1], "/"), ""),
        RatioString: submatches[2],
        Width: width,
        Format: submatches[4],
    }
    return imageReq, nil
}