package physics

import (
	"math"
	"testing"
)

// Тест движения узла под воздействием гравитации
func TestVerletGravity(t *testing.T) {
	w := NewWorld()
	w.Gravity = Vector2{X: 0, Y: 100} // Направлена вниз
	w.Damping = 1.0                 // Без сопротивления воздуха

	node := w.AddNode(0, 0, 1.0, false)

	// Делаем несколько шагов симуляции
	dt := 0.1
	for i := 0; i < 5; i++ {
		w.Update(dt)
	}

	// Точка должна опуститься вниз
	if node.Pos.Y <= 0 {
		t.Errorf("Ожидалось, что узел опустится вниз под силой тяжести, но Y = %f", node.Pos.Y)
	}
}

// Тест упругости связи
func TestSpringConstraint(t *testing.T) {
	w := NewWorld()
	w.Gravity = Vector2{X: 0, Y: 0} // Без гравитации
	w.SubSteps = 16

	na := w.AddNode(0, 0, 1.0, true)
	nb := w.AddNode(100, 0, 1.0, false)

	// Создаем связь с высокой жесткостью 1.0
	// Лимит пластичности 20%, лимит разрыва 50%
	w.AddBeam(na, nb, 1.0, 0.2, 0.5)

	// Силой растягиваем nb до 105 (смещение на 5%, в рамках упругой зоны)
	nb.Pos = Vector2{X: 105, Y: 0}
	nb.OldPos = Vector2{X: 105, Y: 0}

	w.Update(0.1)

	// Связь должна стянуть точку nb обратно ближе к длине покоя 100
	if math.Abs(nb.Pos.X-100) > 1.0 {
		t.Errorf("Ожидалось, что связь вернет точку обратно к 100, но X = %f", nb.Pos.X)
	}
}

// Тест пластической деформации (искривления рамы)
func TestPlasticDeformation(t *testing.T) {
	w := NewWorld()
	w.Gravity = Vector2{X: 0, Y: 0}
	w.SubSteps = 8

	na := w.AddNode(0, 0, 1.0, true)
	nb := w.AddNode(100, 0, 1.0, false)

	// Порог пластичности 10% (10px), лимит разрыва 50% (50px)
	beam := w.AddBeam(na, nb, 1.0, 0.1, 0.5)

	// Растягиваем до 130px (удлинение на 30% превышает лимит пластичности в 10%)
	nb.Pos = Vector2{X: 130, Y: 0}
	nb.OldPos = Vector2{X: 130, Y: 0}

	w.Update(0.016)

	// Длина покоя (RestLength) должна необратимо увеличиться
	if beam.RestLength <= 100.0 {
		t.Errorf("Ожидалось увеличение длины покоя RestLength из-за пластической деформации, но получено %f", beam.RestLength)
	}

	if beam.IsBroken {
		t.Error("Связь не должна была разорваться при растяжении 30% (лимит разрыва 50%)")
	}
}

// Тест разрыва связи
func TestFracture(t *testing.T) {
	w := NewWorld()
	w.Gravity = Vector2{X: 0, Y: 0}

	na := w.AddNode(0, 0, 1.0, true)
	nb := w.AddNode(100, 0, 1.0, false)

	// Лимит разрыва 20% (20px)
	beam := w.AddBeam(na, nb, 1.0, 0.1, 0.2)

	// Растягиваем до 135px (удлинение на 35% превышает лимит разрыва в 20%)
	nb.Pos = Vector2{X: 135, Y: 0}
	nb.OldPos = Vector2{X: 135, Y: 0}

	w.Update(0.016)

	if !beam.IsBroken {
		t.Error("Ожидалось, что связь порвется, так как нагрузка превысила лимит разрыва")
	}
}

// Тест коллизии с холмистым ландшафтом
func TestTerrainCollision(t *testing.T) {
	w := NewWorld()
	w.Gravity = Vector2{X: 0, Y: 0}

	// Ландшафт с наклоном: y = 500 - x
	w.TerrainFunc = func(x float64) float64 {
		return 500.0 - x
	}

	// Узел в точке X=100. Высота земли в этой точке Y = 500 - 100 = 400.
	// Поместим узел в (100, 450), что находится под землей.
	node := w.AddNode(100, 450, 1.0, false)

	w.Update(0.016)

	// Узел должен быть вытолкнут из-под земли на высоту Y <= 400
	if node.Pos.Y > 400.0 {
		t.Errorf("Ожидалось выталкивание узла выше линии земли Y=400, получено Y = %f", node.Pos.Y)
	}
}

// Тест коллизии колеса (узел с радиусом)
func TestWheelCollision(t *testing.T) {
	w := NewWorld()
	w.Gravity = Vector2{X: 0, Y: 100}

	// Плоская земля на Y = 500
	w.TerrainFunc = func(x float64) float64 {
		return 500.0
	}

	// Создаем колесо радиусом 20px на высоте Y = 490 (центр).
	// Нижняя грань колеса находится на Y = 510, что ниже земли (500).
	node := w.AddWheelNode(100, 490, 1.0, 20.0)

	w.Update(0.016)

	// Колесо должно быть вытолкнуто вверх так, чтобы его центр был на Y <= 480 (500 - 20)
	if node.Pos.Y > 480.001 {
		t.Errorf("Ожидалось выталкивание колеса на высоту центра Y=480, получено Y = %f", node.Pos.Y)
	}
}
