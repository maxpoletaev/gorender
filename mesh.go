package main

import (
	"fmt"
	"image/color"
	"os"
	"path"
)

type UV struct {
	U, V float64
}

type Face struct {
	A, B, C       int // Vertex indices
	UVa, UVb, UVc UV  // Texture coordinates
}

type Mesh struct {
	Name        string
	Vertices    []Vec3
	Faces       []Face
	Texture     *Texture
	BoundingBox [8]Vec4
}

func boundingBox(vertices []Vec3) [8]Vec4 {
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

func NewMesh(vertices []Vec3, faces []Face) *Mesh {
	return &Mesh{
		Faces:       faces,
		Vertices:    vertices,
		BoundingBox: boundingBox(vertices),
		Texture: &Texture{
			color: color.RGBA{255, 255, 255, 255},
		},
	}
}

type Object struct {
	*Mesh
	Rotation    Vec3
	Translation Vec3
	Scale       Vec3
}

func NewObject(mesh *Mesh) *Object {
	return &Object{
		Mesh:  mesh,
		Scale: Vec3{1, 1, 1},
	}
}

func LoadMeshFile(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	var mesh *Mesh

	switch ext := path.Ext(filename); ext {
	case ".obj":
		mesh, err = ReadObj(file)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported mesh format: %s", ext)
	}

	mesh.Name = path.Base(filename)

	return mesh, nil
}
