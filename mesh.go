package main

import "image/color"

// UV represents a texture coordinate.
type UV struct {
	U, V float64
}

// Face defines a 3D triangular face in a mesh.
type Face struct {
	A, B, C       int // Vertex indices
	UVa, UVb, UVc UV  // Texture coordinates
}

type Mesh struct {
	Name        string
	Vertices    []Vec3
	Faces       []Face
	Rotation    Vec3
	Translation Vec3
	Scale       Vec3
	Texture     Texture
}

func NewMesh(vertices []Vec3, faces []Face) *Mesh {
	return &Mesh{
		Vertices: vertices,
		Faces:    faces,
		Scale:    Vec3{1, 1, 1},
		Texture: &SolidTexture{
			Color: color.RGBA{255, 255, 255, 255},
		},
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
