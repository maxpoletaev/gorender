package main

import (
	"math"
)

// Vec2 represents a 2D vector.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
}

func (v Vec2) Divide(scalar float64) Vec2 {
	return Vec2{X: v.X / scalar, Y: v.Y / scalar}
}

func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) DotProduct(other Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vec2) Norm() Vec2 {
	return v.Divide(v.Length())
}

// Vec3 represents a 3D vector.
type Vec3 struct {
	X, Y, Z float64
}

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vec3) Divide(scalar float64) Vec3 {
	return Vec3{v.X / scalar, v.Y / scalar, v.Z / scalar}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) RotateX(x float64) Vec3 {
	if x == 0 {
		return v
	}
	return Vec3{
		X: v.X,
		Y: v.Y*math.Cos(x) - v.Z*math.Sin(x),
		Z: v.Y*math.Sin(x) + v.Z*math.Cos(x),
	}
}

func (v Vec3) RotateY(y float64) Vec3 {
	if y == 0 {
		return v
	}
	return Vec3{
		X: v.X*math.Cos(y) + v.Z*math.Sin(y),
		Y: v.Y,
		Z: -v.X*math.Sin(y) + v.Z*math.Cos(y),
	}
}

func (v Vec3) RotateZ(z float64) Vec3 {
	if z == 0 {
		return v
	}
	return Vec3{
		X: v.X*math.Cos(z) - v.Y*math.Sin(z),
		Y: v.X*math.Sin(z) + v.Y*math.Cos(z),
		Z: v.Z,
	}
}

func (v Vec3) Rotate(x, y, z float64) Vec3 {
	return v.RotateX(x).RotateY(y).RotateZ(z)
}

func (v Vec3) CrossProduct(other Vec3) Vec3 {
	x := v.Y*other.Z - v.Z*other.Y
	y := v.Z*other.X - v.X*other.Z
	z := v.X*other.Y - v.Y*other.X
	return Vec3{X: x, Y: y, Z: z}
}

func (v Vec3) DotProduct(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vec3) Normalize() Vec3 {
	return v.Divide(v.Length())
}

func (v Vec3) ToVec4() Vec4 {
	return Vec4{X: v.X, Y: v.Y, Z: v.Z, W: 1}
}

// Vec4 represents a 4D vector.
type Vec4 struct {
	X, Y, Z, W float64
}

func (v Vec4) ToVec3() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.Z}
}

func (v Vec4) Add(other Vec4) Vec4 {
	return Vec4{v.X + other.X, v.Y + other.Y, v.Z + other.Z, v.W + other.W}
}

func (v Vec4) Multiply(scalar float64) Vec4 {
	return Vec4{v.X * scalar, v.Y * scalar, v.Z * scalar, v.W * scalar}
}

func (v Vec4) Divide(scalar float64) Vec4 {
	return Vec4{v.X / scalar, v.Y / scalar, v.Z / scalar, v.W / scalar}
}
