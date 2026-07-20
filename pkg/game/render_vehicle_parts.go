package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"gravitytrails/pkg/physics"
)

func (g *Game) drawDeformableSprite(screen *ebiten.Image) {
	preset := VehiclePresets[g.currentVehicleIndex]
	texture := vehicleTextures[preset.ID]
	if texture == nil {
		return
	}

	v := g.vehicle
	wVal := float32(preset.ChassisWidth)
	hVal := float32(preset.ChassisHeight)

	uvX := []float32{
		wVal * 0.15, wVal * 0.85, wVal, 0, wVal * 0.5,
	}
	uvY := []float32{
		0, 0, hVal, hVal, hVal * 0.5,
	}

	vertices := make([]ebiten.Vertex, 5)
	for i := 0; i < 5; i++ {
		var n *physics.Node
		if i < 4 {
			n = v.Chassis[i]
		} else {
			n = v.Center
		}
		vertices[i] = ebiten.Vertex{
			DstX:   float32(n.Pos.X - g.camX),
			DstY:   float32(n.Pos.Y),
			SrcX:   uvX[i],
			SrcY:   uvY[i],
			ColorR: 1.0, ColorG: 1.0, ColorB: 1.0, ColorA: 1.0,
		}
	}

	var indices []uint16
	if isBeamIntact(v, v.Chassis[0], v.Chassis[1]) && isBeamIntact(v, v.Chassis[0], v.Center) && isBeamIntact(v, v.Chassis[1], v.Center) {
		indices = append(indices, 0, 1, 4)
	}
	if isBeamIntact(v, v.Chassis[1], v.Chassis[2]) && isBeamIntact(v, v.Chassis[1], v.Center) && isBeamIntact(v, v.Chassis[2], v.Center) {
		indices = append(indices, 1, 2, 4)
	}
	if isBeamIntact(v, v.Chassis[2], v.Chassis[3]) && isBeamIntact(v, v.Chassis[2], v.Center) && isBeamIntact(v, v.Chassis[3], v.Center) {
		indices = append(indices, 2, 3, 4)
	}
	if isBeamIntact(v, v.Chassis[3], v.Chassis[0]) && isBeamIntact(v, v.Chassis[3], v.Center) && isBeamIntact(v, v.Chassis[0], v.Center) {
		indices = append(indices, 3, 0, 4)
	}

	if len(indices) > 0 {
		screen.DrawTriangles(vertices, indices, texture, nil)
	}
}

func (g *Game) drawBeams(screen *ebiten.Image) {
	for _, b := range g.world.Beams {
		if b.IsBroken || isChassisBeam(g.vehicle, b) {
			continue
		}

		x1 := float32(b.NodeA.Pos.X - g.camX)
		y1 := float32(b.NodeA.Pos.Y)
		x2 := float32(b.NodeB.Pos.X - g.camX)
		y2 := float32(b.NodeB.Pos.Y)

		strain := b.Stress
		absStrain := math.Abs(strain)

		t := absStrain / b.LimitFracture
		if t > 1.0 {
			t = 1.0
		}

		var c color.RGBA
		if strain > 0 {
			c = color.RGBA{uint8(60 + 195*t), uint8(160 - 120*t), uint8(90 - 40*t), 255}
		} else {
			c = color.RGBA{uint8(60 - 30*t), uint8(160 - 60*t), uint8(90 + 165*t), 255}
		}

		if absStrain > b.LimitPlastic {
			c.R = uint8(math.Min(float64(c.R)*1.2, 255))
			c.G = uint8(math.Min(float64(c.G)*1.2, 255))
			c.B = uint8(math.Min(float64(c.B)*1.2, 255))
		}

		thickness := float32(2.0)
		if strain < 0 {
			thickness += float32(absStrain * 12)
		}

		vector.StrokeLine(screen, x1, y1, x2, y2, thickness, c, true)
	}
}

func (g *Game) drawDetachableParts(screen *ebiten.Image) {
	v := g.vehicle
	preset := VehiclePresets[g.currentVehicleIndex]
	var bodyColor color.Color
	switch preset.ID {
	case "buggy":
		bodyColor = color.RGBA{0, 160, 120, 255}
	case "sport":
		bodyColor = color.RGBA{220, 40, 40, 255}
	case "truck":
		bodyColor = color.RGBA{140, 80, 220, 255}
	case "bike":
		bodyColor = color.RGBA{255, 100, 0, 255}
	default:
		bodyColor = color.RGBA{100, 100, 110, 255}
	}

	if v.Spoiler != nil {
		scx := float32(v.Spoiler.Pos.X - g.camX)
		scy := float32(v.Spoiler.Pos.Y)

		if isBeamIntact(v, v.Spoiler, v.Chassis[0]) {
			x1 := float32(v.Chassis[0].Pos.X - g.camX)
			y1 := float32(v.Chassis[0].Pos.Y)
			vector.StrokeLine(screen, scx, scy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}
		if isBeamIntact(v, v.Spoiler, v.Chassis[3]) {
			x1 := float32(v.Chassis[3].Pos.X - g.camX)
			y1 := float32(v.Chassis[3].Pos.Y)
			vector.StrokeLine(screen, scx, scy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}

		vector.StrokeLine(screen, scx-12, scy, scx+12, scy, 4, color.RGBA{30, 30, 30, 255}, true)
	}

	if v.Bumper != nil {
		bcx := float32(v.Bumper.Pos.X - g.camX)
		bcy := float32(v.Bumper.Pos.Y)

		if isBeamIntact(v, v.Bumper, v.Chassis[1]) {
			x1 := float32(v.Chassis[1].Pos.X - g.camX)
			y1 := float32(v.Chassis[1].Pos.Y)
			vector.StrokeLine(screen, bcx, bcy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}
		if isBeamIntact(v, v.Bumper, v.Chassis[2]) {
			x1 := float32(v.Chassis[2].Pos.X - g.camX)
			y1 := float32(v.Chassis[2].Pos.Y)
			vector.StrokeLine(screen, bcx, bcy, x1, y1, 2, color.RGBA{100, 100, 110, 255}, true)
		}

		vector.StrokeLine(screen, bcx, bcy-10, bcx, bcy+10, 5, bodyColor, true)
	}
}
