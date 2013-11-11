package main

import (
	"net/http"
    "math"

    "github.com/rafikk/imagick/imagick"
)

var backgroundColors = []string{
    "rgb(153,153,51)",
    "rgb(102,153,51)",
    "rgb(51,153,51)",
    "rgb(153,51,51)",
    "rgb(194,133,71)",
    "rgb(51,153,102)",
    "rgb(153,51,102)",
    "rgb(71,133,194)",
    "rgb(51,153,153)",
    "rgb(153,51,153)",
}
var backgroundIndex = 0

func placeholder(w http.ResponseWriter, r BettyRequest) {

    imagick.Initialize()
    defer imagick.Terminate()

    mw := imagick.NewMagickWand()
    dw := imagick.NewDrawingWand()

    if r.RatioString == "original" {
        var height = int(math.Floor(float64(r.Width) * 9.0 / 16.0))
        mw.SetSize(uint(r.Width), uint(height))
    } else {
        mw.SetSize(uint(r.Size().Max.X), uint(r.Size().Max.Y))
    }
    dw.SetFont(config.PlaceholderFont)
    dw.SetFontSize(52)

    pw := imagick.NewPixelWand()
    pw.SetColor("#FFFFFF")
    dw.SetFillColor(pw)
    pw.Destroy()

    var backgroundColor = backgroundColors[backgroundIndex%len(backgroundColors)]
    backgroundIndex += 1

    mw.ReadImage("xc:" + backgroundColor)
    dw.SetGravity(imagick.GRAVITY_CENTER)

    mw.AnnotateImage(dw, 0, 0, 0, r.RatioString)
    mw.DrawImage(dw)

	if r.Format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
        mw.SetFormat("JPEG")
	}
	if r.Format == "png" {
		w.Header().Set("Content-Type", "image/png")
        mw.SetFormat("PNG")
	}
    imageBytes := mw.GetImageBlob()
    mw.Destroy()
    dw.Destroy()
    w.Write(imageBytes)
    return
}
