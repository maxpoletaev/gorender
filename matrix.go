package main

import "math"

type Mat4 [16]float64

func NewIdentity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewScale(x, y, z float64) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func NewTranslation(x, y, z float64) Mat4 {
	return Mat4{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}
}

func NewRotationX(angle float64) Mat4 {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := math.Sin(angle), math.Cos(angle)
	return Mat4{
		1, 0, 0, 0,
		0, cos, -sin, 0,
		0, sin, cos, 0,
		0, 0, 0, 1,
	}
}

func NewRotationY(angle float64) Mat4 {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := math.Sin(angle), math.Cos(angle)
	return Mat4{
		cos, 0, sin, 0,
		0, 1, 0, 0,
		-sin, 0, cos, 0,
		0, 0, 0, 1,
	}
}

func NewRotationZ(angle float64) Mat4 {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := math.Sin(angle), math.Cos(angle)
	return Mat4{
		cos, -sin, 0, 0,
		sin, cos, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewPerspective(fov, aspect, znear, zfar float64) Mat4 {
	f := 1.0 / math.Tan(fov/2.0)
	m := NewIdentity()

	m.set(0, 0, f/aspect)
	m.set(1, 1, f)
	m.set(2, 2, zfar/(znear-zfar))
	m.set(2, 3, (-zfar*znear)/(zfar-znear))
	m.set(3, 2, 1)

	return m
}

func (m *Mat4) at(row, col int) float64 {
	return m[row*4+col]
}

func (m *Mat4) set(row, col int, value float64) {
	m[row*4+col] = value
}

func (m Mat4) Multiply(other Mat4) Mat4 {
	result := Mat4{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i*4+j] = m.at(i, 0)*other.at(0, j) +
				m.at(i, 1)*other.at(1, j) +
				m.at(i, 2)*other.at(2, j) +
				m.at(i, 3)*other.at(3, j)
		}
	}
	return result
}

func (m Mat4) MultiplyVec4(v Vec4) Vec4 {
	return Vec4{
		X: m.at(0, 0)*v.X + m.at(0, 1)*v.Y + m.at(0, 2)*v.Z + m.at(0, 3)*v.W,
		Y: m.at(1, 0)*v.X + m.at(1, 1)*v.Y + m.at(1, 2)*v.Z + m.at(1, 3)*v.W,
		Z: m.at(2, 0)*v.X + m.at(2, 1)*v.Y + m.at(2, 2)*v.Z + m.at(2, 3)*v.W,
		W: m.at(3, 0)*v.X + m.at(3, 1)*v.Y + m.at(3, 2)*v.Z + m.at(3, 3)*v.W,
	}
}
