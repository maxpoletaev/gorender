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
	Name     string
	Vertices []Vec3
	Faces    []Face
	Texture  *Texture
}

func NewMesh(vertices []Vec3, faces []Face) *Mesh {
	return &Mesh{
		Vertices: vertices,
		Faces:    faces,
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
