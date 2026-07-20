package physics

import (
	"math"
)

// Node представляет материальную точку в системе Верле
type Node struct {
	Pos      Vector2 // Текущая позиция
	OldPos   Vector2 // Предыдущая позиция (для расчета скорости)
	Mass     float64 // Масса точки
	IsStatic bool    // Если true, точка закреплена и не двигается
	Force    Vector2 // Накапливаемая сила за кадр
	IsWheel  bool    // Флаг, является ли узел колесом
	Radius   float64 // Радиус колеса (используется для коллизий)
	Rotation float64 // Угол поворота колеса в радианах (для анимации качения)
	OnGround bool    // Флаг касания ландшафта в текущем кадре
	Friction float64 // Индивидуальный коэффициент трения узла при коллизии
}

// Beam представляет упруго-пластичную связь (пружину) между двумя узлами
type Beam struct {
	NodeA         *Node
	NodeB         *Node
	RestLength    float64 // Длина связи в состоянии покоя
	Stiffness     float64 // Жесткость связи (от 0.0 до 1.0)
	LimitPlastic  float64 // Порог пластической деформации (% от RestLength)
	LimitFracture float64 // Порог разрушения/разрыва (% от RestLength)
	IsBroken      bool    // Флаг разрушения
	Stress        float64 // Текущее натяжение для визуализации (-1.0 до 1.0)
}

// World представляет физический мир, содержащий все точки и связи
type World struct {
	Nodes       []*Node
	Beams       []*Beam
	Gravity     Vector2
	Damping     float64                      // Коэффициент затухания скорости (сопротивление воздуха)
	SubSteps    int                          // Количество итераций разрешения связей за шаг
	TerrainFunc func(x float64) float64      // Функция ландшафта: возвращает Y для заданной X
}

// NewWorld создает новый физический мир с дефолтными параметрами
func NewWorld() *World {
	return &World{
		Nodes:    make([]*Node, 0),
		Beams:    make([]*Beam, 0),
		Gravity:  Vector2{X: 0, Y: 400}, // Гравитация направлена вниз
		Damping:  0.99,
		SubSteps: 8,
	}
}

// AddNode добавляет новый узел в физический мир
func (w *World) AddNode(x, y float64, mass float64, isStatic bool) *Node {
	n := &Node{
		Pos:      Vector2{X: x, Y: y},
		OldPos:   Vector2{X: x, Y: y},
		Mass:     mass,
		IsStatic: isStatic,
		Friction: 0.35, // Дефолтное трение для узлов рамы
	}
	w.Nodes = append(w.Nodes, n)
	return n
}

// AddWheelNode добавляет колесо (узел с радиусом) в физический мир
func (w *World) AddWheelNode(x, y float64, mass float64, radius float64) *Node {
	n := &Node{
		Pos:      Vector2{X: x, Y: y},
		OldPos:   Vector2{X: x, Y: y},
		Mass:     mass,
		IsWheel:  true,
		Radius:   radius,
		Friction: 0.05, // По умолчанию колеса катятся свободно (низкое сопротивление)
	}
	w.Nodes = append(w.Nodes, n)
	return n
}

// AddBeam добавляет новую связь между двумя узлами
func (w *World) AddBeam(nodeA, nodeB *Node, stiffness, limitPlastic, limitFracture float64) *Beam {
	dist := nodeA.Pos.Dist(nodeB.Pos)
	b := &Beam{
		NodeA:         nodeA,
		NodeB:         nodeB,
		RestLength:    dist,
		Stiffness:     stiffness,
		LimitPlastic:  limitPlastic,
		LimitFracture: limitFracture,
	}
	w.Beams = append(w.Beams, b)
	return b
}

// Clear очищает физический мир
func (w *World) Clear() {
	w.Nodes = w.Nodes[:0]
	w.Beams = w.Beams[:0]
}

// GetHeight возвращает высоту ландшафта (Y) в точке X
func (w *World) GetHeight(x float64) float64 {
	if w.TerrainFunc != nil {
		return w.TerrainFunc(x)
	}
	return 500.0 // Дефолтный плоский пол
}

// GetNormal возвращает нормаль к поверхности ландшафта в точке X, направленную вверх
func (w *World) GetNormal(x float64) Vector2 {
	dx := 0.1
	y1 := w.GetHeight(x - dx)
	y2 := w.GetHeight(x + dx)

	// Вектор касательной: T = (2*dx, y2 - y1)
	// Вектор нормали, смотрящий вверх (в отрицательную сторону Y): N = (y2 - y1, -2*dx)
	dy := y2 - y1
	lenVal := math.Sqrt(dy*dy + 4*dx*dx)
	if lenVal == 0 {
		return Vector2{X: 0, Y: -1}
	}
	return Vector2{X: dy / lenVal, Y: -2 * dx / lenVal}
}

// Update обновляет физическое состояние на шаг времени dt (в секундах)
func (w *World) Update(dt float64) {
	if dt <= 0 {
		return
	}

	// Сбрасываем флаг OnGround в начале каждого кадра
	for _, node := range w.Nodes {
		node.OnGround = false
	}

	// 1. Интегрирование Верле для всех узлов
	for _, node := range w.Nodes {
		if node.IsStatic {
			continue
		}

		// Сохраняем текущую позицию
		temp := node.Pos

		// Рассчитываем ускорение: a = g + F/m
		acc := w.Gravity.Add(node.Force.Div(node.Mass))

		// Формула Верле с затуханием: pos_new = pos + (pos - pos_old) * damping + acc * dt^2
		velocity := node.Pos.Sub(node.OldPos).Mul(w.Damping)
		node.Pos = node.Pos.Add(velocity).Add(acc.Mul(dt * dt))

		// Сдвигаем старую позицию
		node.OldPos = temp

		// Сбрасываем временные силы
		node.Force = Vector2{}
	}

	// 2. Разрешение ограничений (Constraint Relaxation)
	for step := 0; step < w.SubSteps; step++ {
		// Разрешаем связи (Beams)
		for _, b := range w.Beams {
			if b.IsBroken {
				continue
			}

			dir := b.NodeB.Pos.Sub(b.NodeA.Pos)
			dist := dir.Len()
			if dist == 0 {
				continue
			}

			// Относительное удлинение (strain)
			elongation := dist - b.RestLength
			strain := elongation / b.RestLength

			// Сохраняем коэффициент нагрузки для рендеринга (Stress)
			b.Stress = strain

			// Проверка разрушения (разрыва) связи
			if math.Abs(strain) > b.LimitFracture {
				b.IsBroken = true
				continue
			}

			// Проверка пластической деформации (металл гнется/мнется)
			threshold := b.LimitPlastic * b.RestLength
			if math.Abs(elongation) > threshold {
				// Рассчитываем величину необратимого сдвига длины покоя
				shift := (math.Abs(elongation) - threshold) * math.Copysign(1, elongation)
				// Часть деформации становится пластической (0.8 - скорость деформации)
				b.RestLength += shift * 0.8
			}

			// Расчет корректировки положения узлов для соблюдения длины покоя
			diff := b.RestLength - dist
			// stiffness регулирует, насколько жестко возвращается форма за одну итерацию
			percent := (diff / dist) * b.Stiffness * 0.5
			offset := dir.Mul(percent)

			if !b.NodeA.IsStatic {
				b.NodeA.Pos = b.NodeA.Pos.Sub(offset)
			}
			if !b.NodeB.IsStatic {
				b.NodeB.Pos = b.NodeB.Pos.Add(offset)
			}
		}

		// 3. Разрешение коллизий с холмистым ландшафтом
		for _, node := range w.Nodes {
			if node.IsStatic {
				continue
			}

			groundY := w.GetHeight(node.Pos.X)
			normal := w.GetNormal(node.Pos.X)

			radius := 0.0
			if node.IsWheel {
				radius = node.Radius
			}

			// Расстояние от узла до земли вдоль нормали
			distToGround := (groundY - node.Pos.Y) / math.Abs(normal.Y)

			if distToGround < radius {
				// Узел касается земли
				node.OnGround = true

				// Рассчитываем величину проникновения
				depth := radius - distToGround

				// Выталкиваем узел из земли вдоль нормали
				node.Pos = node.Pos.Add(normal.Mul(depth))

				// Рассчитываем скорость узла: v = Pos - OldPos
				vel := node.Pos.Sub(node.OldPos)

				// Проекция скорости на нормаль
				velNormal := vel.Dot(normal)

				// Обновляем угол вращения колеса на основе горизонтального перемещения
				if node.IsWheel {
					node.Rotation += (node.Pos.X - node.OldPos.X) / node.Radius
				}

				// Если узел движется в сторону земли (внутрь холма)
				if velNormal < 0 {
					// Разделяем скорость на нормальную и тангенциальную (касательную) составляющие
					velTangent := vel.Sub(normal.Mul(velNormal))

					// Коэффициенты упругости и трения
					bounce := 0.1
					// Используем индивидуальное трение узла (например, низкое для свободного качения, высокое для торможения)
					friction := node.Friction

					if node.IsWheel {
						bounce = 0.05
					}

					// Отражаем нормальную скорость (отскок) и гасим тангенциальную (трение)
					subStepFriction := friction / float64(w.SubSteps)
					if subStepFriction > 1.0 {
						subStepFriction = 1.0
					}
					velNormal = -velNormal * bounce
					velTangent = velTangent.Mul(1.0 - subStepFriction)

					// Новая скорость узла
					newVel := velTangent.Add(normal.Mul(velNormal))

					// Корректируем OldPos для фиксации новой скорости в симуляции Верле
					node.OldPos = node.Pos.Sub(newVel)
				}
			}
		}
	}
}
