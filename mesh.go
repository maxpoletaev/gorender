package main

type Triangle struct {
	A, B, C Vec2
	Z       float64
}

type Face struct {
	A, B, C int
}

type Mesh struct {
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

func (m *Mesh) Transform(matrices ...Mat4) {
	mat := matrices[0]
	for i := 1; i < len(matrices); i++ {
		mat = mat.Multiply(matrices[i])
	}
	for j, v := range m.Vertices {
		m.Vertices[j] = mat.MultiplyVec4(v.ToVec4()).ToVec3()
	}
}

// FaceNormal returns the normal vector of a face with the given index.
func (m *Mesh) FaceNormal(idx int) Vec3 {
	a := m.Vertices[m.Faces[idx].A]
	b := m.Vertices[m.Faces[idx].B]
	c := m.Vertices[m.Faces[idx].C]

	return b.Sub(a).CrossProduct(c.Sub(a))
}
