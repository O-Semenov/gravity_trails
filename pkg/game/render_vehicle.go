package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawNodesAndWheels(screen *ebiten.Image) {
	for _, node := range g.world.Nodes {
		cx := float32(node.Pos.X - g.camX)
		cy := float32(node.Pos.Y)

		if node.IsWheel {
			isLeftWheel := node == g.vehicle.LeftWheel
			if g.vehicle.IsWheelBroken(isLeftWheel) {
				vector.DrawFilledCircle(screen, cx, cy, float32(node.Radius), color.RGBA{80, 80, 90, 180}, true)
				continue
			}

			radius := float32(node.Radius)
			rimRadius := radius * 0.7

			compressAmount := 0.0
			if node.OnGround {
				maxLoad := 0.0
				for _, b := range g.vehicle.Beams {
					if (b.NodeA == node || b.NodeB == node) && b.Stiffness < 0.5 {
						load := -b.Stress
						if load > maxLoad {
							maxLoad = load
						}
					}
				}
				compressAmount = math.Min(0.25, math.Max(0.0, maxLoad*0.45))
			}

			normal := g.world.GetNormal(node.Pos.X)
			contactAngle := math.Atan2(-normal.Y, -normal.X)

			ptsX := make([]float32, 24)
			ptsY := make([]float32, 24)

			for i := 0; i < 24; i++ {
				angle := float64(i) * 2.0 * math.Pi / 24.0
				r := radius

				if node.OnGround && compressAmount > 0 {
					diff := math.Abs(angle - contactAngle)
					if diff > math.Pi {
						diff = 2.0*math.Pi - diff
					}
					if diff < math.Pi/3.0 {
						factor := math.Cos(diff * 1.5)
						r = radius * float32(1.0-compressAmount*factor)
					}
				}

				ptsX[i] = cx + float32(math.Cos(angle))*r
				ptsY[i] = cy + float32(math.Sin(angle))*r
			}

			vector.DrawFilledCircle(screen, cx, cy, rimRadius, color.RGBA{35, 35, 45, 255}, true)
			vector.StrokeCircle(screen, cx, cy, rimRadius-1.0, 2, color.RGBA{180, 180, 200, 255}, true)

			for i := 0; i < 4; i++ {
				angle := node.Rotation + float64(i)*math.Pi/2.0
				sx := cx + float32(math.Cos(angle))*(rimRadius-2)
				sy := cy + float32(math.Sin(angle))*(rimRadius-2)
				vector.StrokeLine(screen, cx, cy, sx, sy, 2, color.RGBA{120, 120, 140, 255}, true)
			}

			tireColor := color.RGBA{50, 50, 60, 255}
			for i := 0; i < 24; i++ {
				next := (i + 1) % 24
				vector.StrokeLine(screen, ptsX[i], ptsY[i], ptsX[next], ptsY[next], 4.0, tireColor, true)
			}

			vector.DrawFilledCircle(screen, cx, cy, 5, color.RGBA{0, 255, 230, 255}, true)
			vector.DrawFilledCircle(screen, cx, cy, 3, color.RGBA{20, 20, 30, 255}, true)
		} else {
			isVehiclePart := false
			for _, cn := range g.vehicle.Chassis {
				if node == cn {
					isVehiclePart = true
					break
				}
			}

			preset := VehiclePresets[g.currentVehicleIndex]
			hasTexture := vehicleTextures[preset.ID] != nil
			isCustomBike := preset.ID == "bike"

			if node == g.draggedNode {
				vector.DrawFilledCircle(screen, cx, cy, 5, color.RGBA{0, 255, 230, 200}, true)
			} else if !isVehiclePart || !(hasTexture || isCustomBike) || g.showFrameOnTop {
				var nodeColor color.RGBA
				if node.IsStatic {
					nodeColor = color.RGBA{255, 215, 0, 255}
				} else {
					nodeColor = color.RGBA{200, 200, 210, 255}
				}
				vector.DrawFilledCircle(screen, cx, cy, 7, color.RGBA{30, 30, 40, 255}, true)
				vector.DrawFilledCircle(screen, cx, cy, 4, nodeColor, true)
			}
		}
	}
}
