package main

import (
	"github.com/orsinium-labs/tinymath"
)

type Quaternion struct {
	X, Y, Z, W float32
}

func NewQuaternionFromAxisAngle(axis Vec3, angle float32) Quaternion {
	sin, cos := tinymath.SinCos(angle / 2)
	return Quaternion{
		X: axis.X * sin,
		Y: axis.Y * sin,
		Z: axis.Z * sin,
		W: cos,
	}
}

func (q Quaternion) Multiply(other Quaternion) Quaternion {
	x := q.W*other.X + q.X*other.W + q.Y*other.Z - q.Z*other.Y
	y := q.W*other.Y - q.X*other.Z + q.Y*other.W + q.Z*other.X
	z := q.W*other.Z + q.X*other.Y - q.Y*other.X + q.Z*other.W
	w := q.W*other.W - q.X*other.X - q.Y*other.Y - q.Z*other.Z
	return Quaternion{X: x, Y: y, Z: z, W: w}
}

func (q Quaternion) Conjugate() Quaternion {
	return Quaternion{X: -q.X, Y: -q.Y, Z: -q.Z, W: q.W}
}

func (q Quaternion) Rotate(v Vec3) Vec3 {
	vq := Quaternion{v.X, v.Y, v.Z, 0}
	vr := q.Multiply(vq).Multiply(q.Conjugate())
	return Vec3{vr.X, vr.Y, vr.Z}
}
