package game

import "math"

type GameMap struct {
	ID          string
	Name        string
	Description string
	TerrainFunc func(x float64) float64
}

var GameMaps = []GameMap{
	{
		ID:          "hills",
		Name:        "ХОЛМЫ",
		Description: "Обычный рельеф для тренировки.",
		TerrainFunc: func(x float64) float64 {
			return 450.0 + math.Sin(x*0.005)*90.0 + math.Sin(x*0.02)*18.0 + math.Sin(x*0.08)*4.0
		},
	},
	{
		ID:          "ramps",
		Name:        "ТРАМПЛИНЫ",
		Description: "Разгон и прыжки с трамплинов!",
		TerrainFunc: func(x float64) float64 {
			cycle := 1200.0
			dx := math.Mod(x, cycle)
			if dx < 0 {
				dx += cycle
			}
			baseY := 480.0
			if dx < 400 {
				return baseY
			} else if dx < 550 {
				t := (dx - 400) / 150.0
				return baseY - 120.0*(t*t)
			} else if dx < 580 {
				t := (dx - 550) / 30.0
				return 340.0 + 140.0*t
			} else if dx < 750 {
				t := (dx - 580) / 170.0
				return 480.0 + 80.0*math.Sin(t*math.Pi/2.0)
			} else if dx < 950 {
				t := (dx - 750) / 200.0
				return 560.0 - 100.0*(t*t*(3.0-2.0*t))
			} else {
				return baseY
			}
		},
	},
	{
		ID:          "mountains",
		Name:        "ГОРЫ",
		Description: "Крутые подъемы и спуски.",
		TerrainFunc: func(x float64) float64 {
			return 400.0 + math.Sin(x*0.004)*180.0 + math.Sin(x*0.015)*35.0 + math.Sin(x*0.06)*6.0
		},
	},
}

func getSmoothedHeight(x float64, baseFunc func(x float64) float64) float64 {
	spawnFlatX := 350.0
	spawnTransitionX := 150.0
	spawnBaseY := 480.0

	if x < spawnFlatX {
		return spawnBaseY
	} else if x < spawnFlatX+spawnTransitionX {
		t := (x - spawnFlatX) / spawnTransitionX
		blend := t * t * (3.0 - 2.0*t) // smoothstep
		return spawnBaseY + (baseFunc(x)-spawnBaseY)*blend
	}
	return baseFunc(x)
}
