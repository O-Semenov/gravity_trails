package game

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

//go:embed assets/*
var assetsFS embed.FS

const assetsDir = "assets"

var fontFaceSource *text.GoTextFaceSource
var vehicleTextures = make(map[string]*ebiten.Image)

func init() {
	var err error
	fontFaceSource, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}

	// Загрузка встроенных текстур кузовов при старте (если они есть)
	extensions := []string{".png", ".jpg", ".jpeg"}
	for _, preset := range VehiclePresets {
		vehicleTextures[preset.ID] = nil
		for _, ext := range extensions {
			filePath := assetsDir + "/" + preset.ID + ext
			if file, err := assetsFS.Open(filePath); err == nil {
				img, _, decodeErr := image.Decode(file)
				file.Close()
				if decodeErr == nil {
					vehicleTextures[preset.ID] = ebiten.NewImageFromImage(img)
				}
				break
			}
		}
	}
}


func drawText(screen *ebiten.Image, txt string, x, y float64, size float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.LayoutOptions.LineSpacing = size * 1.2
	if clr != nil {
		op.ColorScale.ScaleWithColor(clr)
	}
	text.Draw(screen, txt, &text.GoTextFace{
		Source: fontFaceSource,
		Size:   size,
	}, op)
}
