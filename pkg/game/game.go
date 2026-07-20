package game

import (
	"gravitytrails/pkg/physics"
	"github.com/hajimehoshi/ebiten/v2"
)

type GameState int

const (
	StateMenu GameState = iota
	StateGame
)

type Game struct {
	screenWidth         int
	screenHeight        int
	world               *physics.World
	vehicle             *Vehicle
	draggedNode         *physics.Node
	camX                float64 // Текущая позиция камеры по горизонтали
	isCrashed           bool    // Флаг аварии
	maxDistance         float64 // Текущий рекорд для активной карты
	allHighScores       map[string]float64
	crashedTimer        float64 // Таймер до фиксации крушения при перевороте (секунды)
	wheelLossTimer      float64 // Таймер до фиксации поражения при отрыве обоих колес
	currentSessionMax   float64 // Максимальная дистанция в текущей сессии
	currentMapIndex     int     // Индекс активной карты
	currentVehicleIndex int     // Индекс активного автомобиля
	state               GameState
	hoveredVehicleIndex int
	hoveredMapIndex     int
}

func NewGame(width, height int) *Game {
	g := &Game{
		screenWidth:         width,
		screenHeight:        height,
		world:               physics.NewWorld(),
		camX:                0,
		currentMapIndex:     0,
		currentVehicleIndex: 0,
		state:               StateMenu,
		hoveredVehicleIndex: -1,
		hoveredMapIndex:     -1,
	}
	g.loadHighScore() // Загрузка рекорда
	g.initSandbox()
	return g
}

func (g *Game) initSandbox() {
	g.world.Clear()
	g.isCrashed = false
	g.crashedTimer = 0.0
	g.wheelLossTimer = 0.0
	g.currentSessionMax = 0.0

	// Настраиваем процедурный ландшафт выбранной карты с плавным спавном
	activeMap := GameMaps[g.currentMapIndex]
	g.world.TerrainFunc = func(x float64) float64 {
		return getSmoothedHeight(x, activeMap.TerrainFunc)
	}

	// Координаты спавна машины
	startX := 150.0
	startY := 200.0

	// Создаем автомобиль выбранного пресета
	g.vehicle = SpawnVehiclePreset(g.world, startX, startY, VehiclePresets[g.currentVehicleIndex])
}

func (g *Game) Update() error {
	if g.state == StateMenu {
		return g.UpdateMenu()
	}
	return g.UpdateGame()
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.state == StateMenu {
		g.DrawMenu(screen)
	} else {
		g.DrawGame(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}
