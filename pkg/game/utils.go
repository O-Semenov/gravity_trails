package game

import (
	"gravitytrails/pkg/physics"
)

func isBeamIntact(v *Vehicle, nodeA, nodeB *physics.Node) bool {
	for _, b := range v.Beams {
		if (b.NodeA == nodeA && b.NodeB == nodeB) || (b.NodeA == nodeB && b.NodeB == nodeA) {
			return !b.IsBroken
		}
	}
	return false
}

func isChassisBeam(v *Vehicle, b *physics.Beam) bool {
	inChassisA := false
	inChassisB := false
	for _, n := range v.Chassis {
		if b.NodeA == n {
			inChassisA = true
		}
		if b.NodeB == n {
			inChassisB = true
		}
	}
	return inChassisA && inChassisB
}

func isVehicleBeam(v *Vehicle, b *physics.Beam) bool {
	for _, vb := range v.Beams {
		if vb == b {
			return true
		}
	}
	return false
}
