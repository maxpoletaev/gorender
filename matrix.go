package main

type Matrix [4][4]float32

func NewIdentityMatrix() Matrix {
	return Matrix{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func NewScaleMatrix(x, y, z float32) Matrix {
	return Matrix{
		{x, 0, 0, 0},
		{0, y, 0, 0},
		{0, 0, z, 0},
		{0, 0, 0, 1},
	}
}

func NewTranslationMatrix(x, y, z float32) Matrix {
	return Matrix{
		{1, 0, 0, x},
		{0, 1, 0, y},
		{0, 0, 1, z},
		{0, 0, 0, 1},
	}
}

func NewRotationXMatrix(angle float32) Matrix {
	if angle == 0 {
		return NewIdentityMatrix()
	}

	sin, cos := sin32(angle), cos32(angle)
	return Matrix{
		{1, 0, 0, 0},
		{0, cos, -sin, 0},
		{0, sin, cos, 0},
		{0, 0, 0, 1},
	}
}

func NewRotationYMatrix(angle float32) Matrix {
	if angle == 0 {
		return NewIdentityMatrix()
	}

	sin, cos := sin32(angle), cos32(angle)
	return Matrix{
		{cos, 0, sin, 0},
		{0, 1, 0, 0},
		{-sin, 0, cos, 0},
		{0, 0, 0, 1},
	}
}

func NewRotationZMatrix(angle float32) Matrix {
	if angle == 0 {
		return NewIdentityMatrix()
	}

	sin, cos := sin32(angle), cos32(angle)
	return Matrix{
		{cos, -sin, 0, 0},
		{sin, cos, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func NewRotationMatrix(x, y, z float32) Matrix {
	m := NewIdentityMatrix()
	m = m.Multiply(NewRotationXMatrix(x))
	m = m.Multiply(NewRotationYMatrix(y))
	m = m.Multiply(NewRotationZMatrix(z))
	return m
}

func NewWorldMatrix(scale, rotation, translation Vec3) Matrix {
	m := NewIdentityMatrix()
	m = NewScaleMatrix(scale.X, scale.Y, scale.Z).Multiply(m)
	m = NewRotationMatrix(rotation.X, rotation.Y, rotation.Z).Multiply(m)
	m = NewTranslationMatrix(translation.X, translation.Y, translation.Z).Multiply(m)
	return m
}

// NewPerspectiveMatrix returns a perspective projection matrix that transforms
// world coordinates to clip coordinates.
func NewPerspectiveMatrix(fov, aspect, zNear, zFar float32) Matrix {
	tanHalfFov := tan32(fov / 2.0)

	m00 := 1 / (aspect * tanHalfFov)
	m11 := 1 / tanHalfFov
	m22 := (zFar + zNear) / (zNear - zFar)
	m23 := (2 * zFar * zNear) / (zNear - zFar)

	return Matrix{
		{m00, 0, 0, 0},
		{0, m11, 0, 0},
		{0, 0, -m22, -m23},
		{0, 0, -1, 0},
	}
}

func NewScreenMatrix(width, height int) Matrix {
	hw := float32(width) / 2
	hh := float32(height) / 2

	return Matrix{
		{hw, 0, 0, hw},
		{0, hh, 0, hh},
		{0, 0, 0.5, 0.5},
		{0, 0, 0, 1},
	}
}

func NewLookAtMatrix(eye, target, up Vec3) Matrix {
	z := target.Sub(eye).Normalize()
	x := up.CrossProduct(z).Normalize()
	y := z.CrossProduct(x).Normalize()

	return Matrix{
		{x.X, x.Y, x.Z, -x.DotProduct(eye)},
		{y.X, y.Y, y.Z, -y.DotProduct(eye)},
		{z.X, z.Y, z.Z, -z.DotProduct(eye)},
		{0, 0, 0, 1},
	}
}

func NewViewMatrix(eye, direction, up Vec3) Matrix {
	z := direction.Normalize()
	x := up.CrossProduct(z).Normalize()
	y := z.CrossProduct(x).Normalize()

	return Matrix{
		{x.X, x.Y, x.Z, -x.DotProduct(eye)},
		{y.X, y.Y, y.Z, -y.DotProduct(eye)},
		{z.X, z.Y, z.Z, -z.DotProduct(eye)},
		{0, 0, 0, 1},
	}
}

func (m Matrix) Multiply(other Matrix) (res Matrix) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				res[i][j] += m[i][k] * other[k][j]
			}
		}
	}

	return res
}

func matrixMultiplyVec4(m *Matrix, v *Vec4) Vec4 {
	return Vec4{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3]*v.W,
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3]*v.W,
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3]*v.W,
		W: m[3][0]*v.X + m[3][1]*v.Y + m[3][2]*v.Z + m[3][3]*v.W,
	}
}
