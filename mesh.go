package main

import "image/color"

type Triangle struct {
	FaceIndex int
	A, B, C   Vec2
	Z         float64
	Color     color.RGBA
}

type UV struct {
	U, V float64
}

type Face struct {
	A, B, C    int // Pointers to the vertices in the mesh.
	TxtCoordsA UV
	TxtCoordsB UV
	TxtCoordsC UV
}

type Mesh struct {
	Name        string
	Vertices    []Vec3
	Faces       []Face
	Rotation    Vec3
	Translation Vec3
	Scale       Vec3
}

func NewMesh(vertices []Vec3, faces []Face) *Mesh {
	return &Mesh{
		Vertices: vertices,
		Faces:    faces,
		Scale:    Vec3{1, 1, 1},
	}
}

func (m *Mesh) Transform(matrices ...Matrix) {
	mat := matrices[0]
	for i := 1; i < len(matrices); i++ {
		mat = mat.Multiply(matrices[i])
	}
	for j, v := range m.Vertices {
		m.Vertices[j] = mat.MultiplyVec4(v.ToVec4()).ToVec3()
	}
}
