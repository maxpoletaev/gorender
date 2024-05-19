package main

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"log"
	"os"
	"path"
	"strings"
)

type ObjMaterial struct {
	Name  string
	MapKd string
}

type ObjContext struct {
	Vertices        []Vec3
	Faces           []Face
	TextureVertices []UV
	Textures        map[string]*Texture
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
			&face.A, &face.UVa,
			&face.B, &face.UVc,
			&face.C, &face.UVb,
		)

	case strings.Contains(line, "/") && strings.Count(line, "/") == 6:
		var vtA, vtB, vtC int
		_, err = fmt.Sscanf(
			line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
			&face.A, &vtA, &discard,
			&face.B, &vtB, &discard,
			&face.C, &vtC, &discard,
		)

		face.UVa = c.TextureVertices[vtA-1]
		face.UVb = c.TextureVertices[vtB-1]
		face.UVc = c.TextureVertices[vtC-1]

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

func parseMtlLibFile(filename string) ([]ObjMaterial, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	var materials []ObjMaterial
	var mat *ObjMaterial

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "newmtl "):
			if mat != nil {
				materials = append(materials, *mat)
			}
			name := strings.TrimPrefix(line, "newmtl ")
			mat = &ObjMaterial{Name: name}
		case strings.HasPrefix(line, "map_Kd "):
			mapKd := strings.TrimPrefix(line, "map_Kd ")
			mat.MapKd = mapKd
		}
	}

	if mat != nil {
		materials = append(materials, *mat)
	}

	return materials, nil
}

// LoadObjFile reads a mesh from an .obj file.
// Format description: https://people.computing.clemson.edu/~dhouse/courses/405/docs/brief-obj-file-format.html
func LoadObjFile(filename string) (*Mesh, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	dirname := path.Dir(filename)
	scanner := bufio.NewScanner(file)
	defaultTexture := &Texture{color: color.RGBA{255, 0, 255, 255}}
	var currentTexture *Texture

	c := &ObjContext{Textures: make(map[string]*Texture)}
	textureFiles := make(map[string]*Texture)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "mtllib "):
			mtlLibFile := strings.TrimPrefix(line, "mtllib ")
			log.Printf("[INFO] found mtllib file: %s", mtlLibFile)

			materials, err := parseMtlLibFile(path.Join(dirname, mtlLibFile))
			if err != nil {
				return nil, fmt.Errorf("failed to parse material library: %s", err)
			}

			for _, m := range materials {
				if m.MapKd == "" {
					log.Printf("[INFO] using default texture for material: %s", m.Name)
					c.Textures[m.Name] = defaultTexture
				} else {
					if texture, ok := textureFiles[m.MapKd]; ok {
						c.Textures[m.Name] = texture
					} else {
						log.Printf("[INFO] loading texture: %s", m.MapKd)
						texture, err = LoadTextureFile(path.Join(dirname, m.MapKd))
						if err != nil {
							return nil, fmt.Errorf("failed to load texture: %s", err)
						}

						textureFiles[m.MapKd] = texture
						c.Textures[m.Name] = texture
					}
				}
			}

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

		case strings.HasPrefix(line, "usemtl "):
			mtlName := strings.TrimPrefix(line, "usemtl ")
			currentTexture = c.Textures[mtlName]

		case strings.HasPrefix(line, "f "):
			f, err := parseFace(c, line)
			if err != nil {
				return nil, err
			}
			f.Texture = currentTexture
			c.Faces = append(c.Faces, f)
		}
	}

	return NewMesh(c.Vertices, c.Faces), nil
}
