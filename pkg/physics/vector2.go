package physics

import "math"

type Vector2 struct {
	X, Y float64
}

func (v Vector2) Add(o Vector2) Vector2 {
	return Vector2{v.X + o.X, v.Y + o.Y}
}

func (v Vector2) Sub(o Vector2) Vector2 {
	return Vector2{v.X - o.X, v.Y - o.Y}
}

func (v Vector2) Mul(scalar float64) Vector2 {
	return Vector2{v.X * scalar, v.Y * scalar}
}

func (v Vector2) Div(scalar float64) Vector2 {
	if scalar == 0 {
		return Vector2{}
	}
	return Vector2{v.X / scalar, v.Y / scalar}
}

func (v Vector2) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2) LenSq() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vector2) Dist(o Vector2) float64 {
	dx := v.X - o.X
	dy := v.Y - o.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v Vector2) DistSq(o Vector2) float64 {
	dx := v.X - o.X
	dy := v.Y - o.Y
	return dx*dx + dy*dy
}

func (v Vector2) Normalize() Vector2 {
	l := v.Len()
	if l == 0 {
		return Vector2{}
	}
	return Vector2{v.X / l, v.Y / l}
}

func (v Vector2) Dot(o Vector2) float64 {
	return v.X*o.X + v.Y*o.Y
}
