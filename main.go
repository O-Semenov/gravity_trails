package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gravitytrails/pkg/game"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gravity Trails: Soft-Body Climber")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := game.NewGame(screenWidth, screenHeight)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
