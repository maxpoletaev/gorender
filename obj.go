package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ObjContext struct {
	Vertices        []Vec3
	Faces           []Face
	TextureVertices []UV
}

func parseVertex(line string) (Vec3, error) {
	var x, y, z float64
	_, err := fmt.Sscanf(line, "v %f %f %f", &x, &y, &z)
	return Vec3{x, y, z}, err
}

func parseTextureVertex(line string) (UV, error) {
	var x, y float64
	_, err := fmt.Sscanf(line, "vt %f %f", &x, &y)
	return UV{x, y}, err
}

func parseFace(c *ObjContext, line string) (Face, error) {
	if strings.Count(line, " ") != 3 {
		return Face{}, errors.New("mesh is not triangulated")
	}

	var (
		err     error
		face    Face
		discard int
	)

	switch {
	case strings.Contains(line, "//"):
		_, err = fmt.Sscanf(
			line, "f %d//%d %d//%d %d//%d",
			&face.A, &discard,
			&face.B, &discard,
			&face.C, &discard,
		)

	case strings.Contains(line, "/") && strings.Count(line, "/") == 3:
		_, err = fmt.Sscanf(
			line, "f %d/%d %d/%d %d/%d",
			&face.A, &face.TxtCoordsA,
			&face.B, &face.TxtCoordsB,
			&face.C, &face.TxtCoordsC,
		)

	case strings.Contains(line, "/") && strings.Count(line, "/") == 6:
		var vt0, vt1, vt2 int
		_, err = fmt.Sscanf(
			line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
			&face.A, &vt0, &discard,
			&face.B, &vt1, &discard,
			&face.C, &vt2, &discard,
		)

		face.TxtCoordsA = c.TextureVertices[vt0-1]
		face.TxtCoordsB = c.TextureVertices[vt1-1]
		face.TxtCoordsC = c.TextureVertices[vt2-1]

	default:
		_, err = fmt.Sscanf(
			line, "f %d %d %d",
			&face.A,
			&face.B,
			&face.C,
		)
	}

	// Indices are 1-based in .obj files.
	face.A -= 1
	face.B -= 1
	face.C -= 1

	return face, err
}

// ReadObj reads a mesh from an .obj file.
// Format description: https://people.computing.clemson.edu/~dhouse/courses/405/docs/brief-obj-file-format.html
func ReadObj(reader io.Reader) (*Mesh, error) {
	scanner := bufio.NewScanner(reader)
	c := &ObjContext{}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "v "):
			v, err := parseVertex(line)
			if err != nil {
				return nil, err
			}
			c.Vertices = append(c.Vertices, v)

		case strings.HasPrefix(line, "vt "):
			vt, err := parseTextureVertex(line)
			if err != nil {
				return nil, err
			}
			c.TextureVertices = append(c.TextureVertices, vt)

		case strings.HasPrefix(line, "f "):
			f, err := parseFace(c, line)
			if err != nil {
				return nil, err
			}
			c.Faces = append(c.Faces, f)
		}
	}

	return NewMesh(c.Vertices, c.Faces), nil
}
