package main

import (
	"fmt"
	"path"
)

type UV struct {
	U, V float32
}

type Face struct {
	VertexIndices [3]int
	NormalIndices [3]int
	UVs           [3]UV
	Texture       *Texture
}

type Mesh struct {
	Name          string
	Vertices      []Vec4
	VertexNormals []Vec4
	FaceNormals   []Vec4
	BoundingBox   [8]Vec4
	Faces         []Face
}

func boundingBox(vertices []Vec4) [8]Vec4 {
	minX, minY, minZ := vertices[0].X, vertices[0].Y, vertices[0].Z
	maxX, maxY, maxZ := minX, minY, minZ

	for _, v := range vertices {
		minX = min(minX, v.X)
		minY = min(minY, v.Y)
		minZ = min(minZ, v.Z)
		maxX = max(maxX, v.X)
		maxY = max(maxY, v.Y)
		maxZ = max(maxZ, v.Z)
	}

	return [8]Vec4{
		{minX, minY, minZ, 1},
		{minX, minY, maxZ, 1},
		{minX, maxY, minZ, 1},
		{minX, maxY, maxZ, 1},
		{maxX, minY, minZ, 1},
		{maxX, minY, maxZ, 1},
		{maxX, maxY, minZ, 1},
		{maxX, maxY, maxZ, 1},
	}
}

func NewMesh(vertices []Vec4, vertexNormals []Vec4, faces []Face) *Mesh {
	faceNormals := make([]Vec4, len(faces))
	for i := range faces {
		v0 := vertices[faces[i].VertexIndices[0]].ToVec3()
		v1 := vertices[faces[i].VertexIndices[1]].ToVec3()
		v2 := vertices[faces[i].VertexIndices[2]].ToVec3()
		faceNormals[i] = v1.Sub(v0).CrossProduct(v2.Sub(v0)).Normalize().ToVec4()
	}

	return &Mesh{
		Faces:         faces,
		Vertices:      vertices,
		VertexNormals: vertexNormals,
		FaceNormals:   faceNormals,
		BoundingBox:   boundingBox(vertices),
	}
}

type Object struct {
	*Mesh
	Rotation            Vec3
	Translation         Vec3
	Scale               Vec3
	TransformedVertices []Vec4
	WorldVertexNormals  []Vec4
	WorldFaceNormals    []Vec4
}

func NewObject(mesh *Mesh) *Object {
	return &Object{
		Mesh:                mesh,
		Scale:               Vec3{1, 1, 1},
		TransformedVertices: make([]Vec4, len(mesh.Vertices)),
		WorldFaceNormals:    make([]Vec4, len(mesh.FaceNormals)),
		WorldVertexNormals:  make([]Vec4, len(mesh.VertexNormals)),
	}
}

func LoadMeshFile(filename string) (*Mesh, error) {
	var (
		mesh *Mesh
		err  error
	)

	switch ext := path.Ext(filename); ext {
	case ".obj":
		mesh, err = LoadObjFile(filename)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported mesh format: %s", ext)
	}

	mesh.Name = path.Base(filename)

	return mesh, nil
}
