package main

import (
	"net/http"

    "github.com/rafikk/imagick/imagick"
)

// var backgroundColors = []color.RGBA{
// 	color.RGBA{153, 153, 51, 255},
// 	color.RGBA{102, 153, 51, 255},
// 	color.RGBA{51, 153, 51, 255},
// 	color.RGBA{153, 51, 51, 255},
// 	color.RGBA{194, 133, 71, 255},
// 	color.RGBA{51, 153, 102, 255},
// 	color.RGBA{153, 51, 102, 255},
// 	color.RGBA{71, 133, 194, 255},
// 	color.RGBA{51, 153, 153, 255},
// 	color.RGBA{153, 51, 153, 255},
// }
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

    mw.SetSize(uint(r.Size().Max.X), uint(r.Size().Max.Y))
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
