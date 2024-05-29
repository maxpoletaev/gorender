//go:build amd64 && !purego

package main

//go:noescape
func _matrixMultiplyVec4SSE(mat *Matrix, vecs []Vec4)

func matrixMultiplyVec4Batch(m *Matrix, vecs []Vec4) {
	mat := (*m).Transpose() // SSE is column-major
	_matrixMultiplyVec4SSE(&mat, vecs)
}
