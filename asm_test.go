package main

import (
	"testing"
)

var (
	benchResultVec4 []Vec4
)

func BenchmarkMatrixMultiplyVec4Batch(b *testing.B) {
	m := NewIdentityMatrix()
	m = NewRotationMatrix(0.1, 0.2, 0.3).Multiply(m)
	m = NewTranslationMatrix(1, 2, 3).Multiply(m)

	vecs := make([]Vec4, 1000)
	for i := range vecs {
		vecs[i] = Vec4{
			float32(i),
			float32(i),
			float32(i),
			float32(i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matrixMultiplyVec4Batch(&m, vecs)
	}

	benchResultVec4 = vecs
}
