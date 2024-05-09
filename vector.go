package main

// Vec2 represents a 2D vector.
type Vec2 struct {
	X, Y float32
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v.X + other.X, v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{v.X - other.X, v.Y - other.Y}
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

func (v Vec2) Norm() Vec2 {
	return v.Divide(v.Length())
}

// Vec3 represents a 3D vector.
type Vec3 struct {
	X, Y, Z float32
}

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vec3) Divide(scalar float32) Vec3 {
	return Vec3{v.X / scalar, v.Y / scalar, v.Z / scalar}
}

func (v Vec3) Length() float32 {
	return sqrt32(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) RotateX(x float32) Vec3 {
	if x == 0 {
		return v
	}
	return Vec3{
		X: v.X,
		Y: v.Y*cos32(x) - v.Z*sin32(x),
		Z: v.Y*sin32(x) + v.Z*cos32(x),
	}
}

func (v Vec3) RotateY(y float32) Vec3 {
	if y == 0 {
		return v
	}
	return Vec3{
		X: v.X*cos32(y) + v.Z*sin32(y),
		Y: v.Y,
		Z: -v.X*sin32(y) + v.Z*cos32(y),
	}
}

func (v Vec3) RotateZ(z float32) Vec3 {
	if z == 0 {
		return v
	}
	return Vec3{
		X: v.X*cos32(z) - v.Y*sin32(z),
		Y: v.X*sin32(z) + v.Y*cos32(z),
		Z: v.Z,
	}
}

func (v Vec3) Rotate(x, y, z float32) Vec3 {
	return v.RotateX(x).RotateY(y).RotateZ(z)
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

func (v Vec3) ToVec4() Vec4 {
	return Vec4{X: v.X, Y: v.Y, Z: v.Z, W: 1}
}

type Vec4 struct {
	X, Y, Z, W float32
}

func (v Vec4) ToVec3() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.Z}
}
