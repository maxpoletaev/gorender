package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func parseObjVertex(line string) (Vec3, error) {
	var x, y, z float64
	_, err := fmt.Sscanf(line, "v %f %f %f", &x, &y, &z)
	return Vec3{x, y, z}, err
}

func parseObjFace(line string) (Face, error) {
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
	case strings.Contains(line, "/"):
		_, err = fmt.Sscanf(
			line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
			&face.A, &discard, &discard,
			&face.B, &discard, &discard,
			&face.C, &discard, &discard,
		)
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
	vertices := make([]Vec3, 0)
	faces := make([]Face, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case 'v':
			v, err := parseObjVertex(line)
			if err != nil {
				return nil, err
			}
			vertices = append(vertices, v)
		case 'f':
			f, err := parseObjFace(line)
			if err != nil {
				return nil, err
			}
			faces = append(faces, f)
		}
	}

	return NewMesh(vertices, faces), nil
}
