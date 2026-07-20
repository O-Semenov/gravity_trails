package game

import (
	"gravitytrails/pkg/physics"
)

// ApplyDriveTorque применяет крутящий момент к колесам и реактивный момент к кузову
func (v *Vehicle) ApplyDriveTorque(w *physics.World, torque float64) {
	// Двигатель работает только при контакте колес с ландшафтом
	onGround := v.LeftWheel.OnGround || v.RightWheel.OnGround
	if !onGround {
		return
	}

	// Находим центр масс рамы автомобиля для распределения реактивного момента
	center := physics.Vector2{}
	count := 0.0
	for _, node := range v.Chassis {
		center = center.Add(node.Pos)
		count++
	}
	hasChassis := count > 0
	if hasChassis {
		center = center.Div(count)
	}

	for _, wheel := range []*physics.Node{v.LeftWheel, v.RightWheel} {
		if wheel.OnGround {
			normal := w.GetNormal(wheel.Pos.X)
			tangent := physics.Vector2{X: -normal.Y, Y: normal.X}

			// Сила тяги, прикладываемая к колесу (двигатель толкает колесо вперед)
			driveForce := tangent.Mul(torque * wheel.Mass)
			wheel.Force = wheel.Force.Add(driveForce)

			// Реактивный вращающий момент на кузов (противоположный крутящему моменту)
			if hasChassis {
				// Учитываем индивидуальный коэффициент DriveReaction
				reactionTorque := -torque * v.DriveReaction
				for _, node := range v.Chassis {
					r := node.Pos.Sub(center)
					dist := r.Len()
					if dist == 0 {
						continue
					}

					perp := physics.Vector2{X: -r.Y, Y: r.X}.Normalize()
					rotForce := perp.Mul(reactionTorque * node.Mass)
					node.Force = node.Force.Add(rotForce)
				}
			}
		}
	}
}

// ApplyAirControl применяет вращающий момент в воздухе для контроля наклона (тангажа) автомобиля
func (v *Vehicle) ApplyAirControl(torque float64) {
	// Воздушный контроль активен только когда оба колеса оторвались от земли
	if v.LeftWheel.OnGround || v.RightWheel.OnGround {
		return
	}

	// Находим центр масс рамы автомобиля
	center := physics.Vector2{}
	count := 0.0
	for _, node := range v.Chassis {
		center = center.Add(node.Pos)
		count++
	}
	if count == 0 {
		return
	}
	center = center.Div(count)

	// Распределяем вращающий момент в виде касательных сил на все узлы рамы
	for _, node := range v.Chassis {
		r := node.Pos.Sub(center)
		dist := r.Len()
		if dist == 0 {
			continue
		}

		perp := physics.Vector2{X: -r.Y, Y: r.X}.Normalize()
		rotForce := perp.Mul(torque * node.Mass)
		node.Force = node.Force.Add(rotForce)
	}
}

// GetCenterOfMass возвращает центр масс автомобиля
func (v *Vehicle) GetCenterOfMass() physics.Vector2 {
	center := physics.Vector2{}
	for _, node := range v.Chassis {
		center = center.Add(node.Pos)
	}
	center = center.Add(v.LeftWheel.Pos)
	center = center.Add(v.RightWheel.Pos)
	return center.Div(float64(len(v.Chassis) + 2))
}

// GetVelocity возвращает усредненный вектор скорости автомобиля
func (v *Vehicle) GetVelocity() physics.Vector2 {
	vel := physics.Vector2{}
	for _, node := range v.Chassis {
		vel = vel.Add(node.Pos.Sub(node.OldPos))
	}
	vel = vel.Add(v.LeftWheel.Pos.Sub(v.LeftWheel.OldPos))
	vel = vel.Add(v.RightWheel.Pos.Sub(v.RightWheel.OldPos))
	return vel.Div(float64(len(v.Chassis) + 2))
}

// IsWheelBroken проверяет, оторвано ли конкретное колесо (разорвана ли его подвеска)
func (v *Vehicle) IsWheelBroken(left bool) bool {
	var wheel *physics.Node
	if left {
		wheel = v.LeftWheel
	} else {
		wheel = v.RightWheel
	}

	suspensionCount := 0
	for _, b := range v.Beams {
		if b.IsBroken {
			continue
		}

		var otherNode *physics.Node
		if b.NodeA == wheel {
			otherNode = b.NodeB
		} else if b.NodeB == wheel {
			otherNode = b.NodeA
		}

		if otherNode != nil {
			for _, chNode := range v.Chassis {
				if otherNode == chNode {
					suspensionCount++
					break
				}
			}
		}
	}
	return suspensionCount == 0
}
