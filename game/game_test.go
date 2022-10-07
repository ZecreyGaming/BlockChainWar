package game

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/COAOX/zecrey_warrior/config"
	"github.com/COAOX/zecrey_warrior/db"
)

var img = image.NewRGBA(image.Rect(0, 0, 852, 642))
var col color.Color

// HLine draws a horizontal line
func HLine(x1, y, x2 int) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

// VLine draws a veritcal line
func VLine(x, y1, y2 int) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

// Rect draws a rectangle utilizing HLine() and VLine()
func Rect(x1, y1, x2, y2 int) {
	HLine(x1, y1, x2)
	HLine(x1, y2, x2)
	VLine(x1, y1, y2)
	VLine(x2, y1, y2)
}

func TestGame(t *testing.T) {
	cfg := config.Read("../config/local.json")
	d := db.NewClient(cfg.Database)
	g := NewGame(context.Background(), cfg, d, func(winner Camp) {}, func(camp Camp, votes int32) {})

	new_png_file := "draw.png" // output image will live here

	myimage := image.NewRGBA(image.Rect(0, 0, 852, 642)) // x1,y1,  x2,y2 of background rectangle
	mygreen := color.RGBA{0, 100, 0, 255}                //  R, G, B, Alpha

	// backfill entire background surface with color mygreen
	draw.Draw(myimage, myimage.Bounds(), &image.Uniform{mygreen}, image.ZP, draw.Src)

	// red_rect := image.Rect(60, 80, 120, 160) //  geometry of 2nd rectangle which we draw atop above rectangle
	myred := color.RGBA{200, 0, 0, 255}

	for _, obj := range g.space.Objects() {
		draw.Draw(myimage, image.Rect(int(obj.X), int(obj.Y), int(obj.X+obj.W), int(obj.Y+obj.H)), &image.Uniform{myred}, image.ZP, draw.Src)
	}

	myfile, err := os.Create(new_png_file) // ... now lets save output image
	if err != nil {
		panic(err)
	}
	defer myfile.Close()
	png.Encode(myfile, myimage)
}
