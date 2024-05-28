//go:build !amd64 || noasm

package main

func init() {
	println("noasm")
}

func matrixMultiplyVec4Batch(m *Matrix, vecs []Vec4) {
	for i := range vecs {
		v := &vecs[i]
		vecs[i] = Vec4{
			X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3]*v.W,
			Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3]*v.W,
			Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3]*v.W,
			W: m[3][0]*v.X + m[3][1]*v.Y + m[3][2]*v.Z + m[3][3]*v.W,
		}
	}
}
