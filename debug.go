package main

import (
    "image/color"
    "image/draw"
    "image/jpeg"
    "image/png"
    "image"
    "log"
    "math"
    "net/http"

    "code.google.com/p/freetype-go/freetype"
)


var backgroundColors = []color.RGBA{
    color.RGBA{153, 153, 51, 255},
    color.RGBA{102, 153, 51, 255},
    color.RGBA{51, 153, 51, 255},
    color.RGBA{153, 51, 51, 255},
    color.RGBA{194, 133, 71, 255},
    color.RGBA{51, 153, 102, 255},
    color.RGBA{153, 51, 102, 255},
    color.RGBA{71, 133, 194, 255},
    color.RGBA{51, 153, 153, 255},
    color.RGBA{153, 51, 153, 255},
}
var backgroundIndex = 0

func placeholder(w http.ResponseWriter, r BettyRequest) {
    // TODO: Don't do so much stupid shit with this font stuff.

    var dst = image.NewRGBA(r.Size())
    backgroundIndex += 1
    var backgroundColor = backgroundColors[backgroundIndex%len(backgroundColors)]
    draw.Draw(dst, dst.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)

    font, err := freetype.ParseFont(font_ttf())
    if err != nil {
        log.Println(err)
        return
    }

    var txtImage = image.NewRGBA(image.Rect(0, 0, 600, 600))
    draw.Draw(txtImage, txtImage.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)

    txtColor := image.White

    var fontSize = float64(r.Width) * 52 / 300 // Stupid magic number

    c := freetype.NewContext()
    c.SetDPI(72)
    c.SetFont(font)
    c.SetFontSize(fontSize)
    c.SetClip(txtImage.Bounds())
    c.SetDst(txtImage)
    c.SetSrc(txtColor)

    var offsetFix = int(math.Floor(fontSize * 12 / 72)) // Stupid magic number

    pt := freetype.Pt(0, int(c.PointToFix32(fontSize)>>8)-offsetFix)
    pt, err = c.DrawString(r.RatioString, pt)
    if err != nil {
        log.Println(err)
        return
    }

    txtSize := image.Pt(int(pt.X>>8), int(c.PointToFix32(fontSize)>>8)+2)

    txtBounds := image.Rect(
        int(math.Floor(float64(r.Size().Max.X)/2.0)-(float64(txtSize.X)/2.0)),
        int(math.Floor(float64(r.Size().Max.Y)/2.0)-(float64(txtSize.Y)/2.0)),
        int(math.Floor(float64(r.Size().Max.X)/2.0)+(float64(txtSize.X)/2.0)),
        int(math.Floor(float64(r.Size().Max.Y)/2.0)+(float64(txtSize.Y)/2.0)),
    )

    draw.Draw(dst, txtBounds, txtImage, image.ZP, draw.Src)

    if r.Format == "jpg" {
        w.Header().Set("Content-Type", "image/jpeg")
        jpeg.Encode(w, dst, &jpeg.Options{90})
        return
    }
    if r.Format == "png" {
        w.Header().Set("Content-Type", "image/png")
        png.Encode(w, dst)
        return
    }
}

