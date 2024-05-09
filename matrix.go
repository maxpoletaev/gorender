package main

import "math"

type Matrix [4][4]float64

func NewIdentity() Matrix {
	return Matrix{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func NewScale(x, y, z float64) Matrix {
	return Matrix{
		{x, 0, 0, 0},
		{0, y, 0, 0},
		{0, 0, z, 0},
		{0, 0, 0, 1},
	}
}

func NewTranslation(x, y, z float64) Matrix {
	return Matrix{
		{1, 0, 0, x},
		{0, 1, 0, y},
		{0, 0, 1, z},
		{0, 0, 0, 1},
	}
}

func NewRotationX(angle float64) Matrix {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := math.Sin(angle), math.Cos(angle)
	return Matrix{
		{1, 0, 0, 0},
		{0, cos, -sin, 0},
		{0, sin, cos, 0},
		{0, 0, 0, 1},
	}
}

func NewRotationY(angle float64) Matrix {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := math.Sin(angle), math.Cos(angle)
	return Matrix{
		{cos, 0, sin, 0},
		{0, 1, 0, 0},
		{-sin, 0, cos, 0},
		{0, 0, 0, 1},
	}
}

func NewRotationZ(angle float64) Matrix {
	if angle == 0 {
		return NewIdentity()
	}

	sin, cos := math.Sin(angle), math.Cos(angle)
	return Matrix{
		{cos, -sin, 0, 0},
		{sin, cos, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func NewPerspective(fovRad, aspect, znear, zfar float64) Matrix {
	f := 1.0 / math.Tan(fovRad/2.0)
	m := NewIdentity()

	m[0][0] = f / aspect
	m[1][1] = f
	m[2][2] = zfar / (znear - zfar)
	m[2][3] = (-zfar * znear) / (zfar - znear)
	m[3][2] = 1

	return m
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

func (m Matrix) MultiplyVec4(v Vec4) Vec4 {
	return Vec4{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3]*v.W,
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3]*v.W,
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3]*v.W,
		W: m[3][0]*v.X + m[3][1]*v.Y + m[3][2]*v.Z + m[3][3]*v.W,
	}
}
