package main

type Mat4 [16]float32

func NewIdentity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func NewScale(x, y, z float32) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

func NewTranslation(x, y, z float32) Mat4 {
	return Mat4{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}
}

func NewRotationX(angle float32) Mat4 {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := sin32(angle), cos32(angle)
	return Mat4{
		1, 0, 0, 0,
		0, cos, -sin, 0,
		0, sin, cos, 0,
		0, 0, 0, 1,
	}
}

func NewRotationY(angle float32) Mat4 {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := sin32(angle), cos32(angle)
	return Mat4{
		cos, 0, sin, 0,
		0, 1, 0, 0,
		-sin, 0, cos, 0,
		0, 0, 0, 1,
	}
}

func NewRotationZ(angle float32) Mat4 {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := sin32(angle), cos32(angle)
	return Mat4{
		cos, -sin, 0, 0,
		sin, cos, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func (m *Mat4) at(row, col int) float32 {
	return m[row*4+col]
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
