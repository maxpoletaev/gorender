package main

type Vec2 struct {
	X, Y float32
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
}

func (v Vec2) Multiply(scalar float32) Vec2 {
	return Vec2{X: v.X * scalar, Y: v.Y * scalar}
}

func (v Vec2) Divide(scalar float32) Vec2 {
	return Vec2{X: v.X / scalar, Y: v.Y / scalar}
}

func (v Vec2) Length() float32 {
	return sqrt32(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) DotProduct(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vec2) Normalize() Vec2 {
	return v.Divide(v.Length())
}

type Vec3 struct {
	X, Y, Z float32
}

func Vec3FromArray(arr [3]float32) Vec3 {
	return Vec3{arr[0], arr[1], arr[2]}
}

func (v Vec3) ToVec4() Vec4 {
	return Vec4{v.X, v.Y, v.Z, 1}
}

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vec3) Multiply(scalar float32) Vec3 {
	return Vec3{v.X * scalar, v.Y * scalar, v.Z * scalar}
}

func (v Vec3) Divide(scalar float32) Vec3 {
	return Vec3{v.X / scalar, v.Y / scalar, v.Z / scalar}
}

func (v Vec3) Length() float32 {
	return sqrt32(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) CrossProduct(other Vec3) Vec3 {
	x := v.Y*other.Z - v.Z*other.Y
	y := v.Z*other.X - v.X*other.Z
	z := v.X*other.Y - v.Y*other.X
	return Vec3{X: x, Y: y, Z: z}
}

func (v Vec3) DotProduct(other Vec3) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vec3) Normalize() Vec3 {
	return v.Divide(v.Length())
}

func (v Vec3) ToRadians() Vec3 {
	f := pi32 / 180
	return Vec3{v.X * f, v.Y * f, v.Z * f}
}

type Vec4 struct {
	X, Y, Z, W float32
}

func (v Vec4) ToVec3() Vec3 {
	return Vec3{v.X, v.Y, v.Z}
}

func (v Vec4) Add(other Vec4) Vec4 {
	return Vec4{v.X + other.X, v.Y + other.Y, v.Z + other.Z, v.W + other.W}
}

func (v Vec4) Sub(other Vec4) Vec4 {
	return Vec4{v.X - other.X, v.Y - other.Y, v.Z - other.Z, v.W - other.W}
}

func (v Vec4) Divide(scalar float32) Vec4 {
	return Vec4{X: v.X / scalar, Y: v.Y / scalar, Z: v.Z / scalar, W: v.W / scalar}
}

func (v Vec4) Multiply(scalar float32) Vec4 {
	return Vec4{X: v.X * scalar, Y: v.Y * scalar, Z: v.Z * scalar, W: v.W * scalar}
}

func (v Vec4) DotProduct(other Vec4) float32 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z + v.W*other.W
}

func (v Vec4) Conjugate() Vec4 {
	return Vec4{X: -v.X, Y: -v.Y, Z: -v.Z, W: v.W}
}

func (v Vec4) Length() float32 {
	return sqrt32(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)
}

func (v Vec4) Normalize() Vec4 {
	return v.Divide(v.Length())
}
