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
	Vertices        []Vec4
	Faces           []Face
	TextureVertices []UV
	VertexNormals   []Vec3
	Textures        map[string]*Texture
}

func parseVertex(line string) (Vec4, error) {
	var x, y, z float32
	_, err := fmt.Sscanf(line, "v %f %f %f", &x, &y, &z)
	return Vec4{x, y, z, 1}, err
}

func parseTextureVertex(line string) (UV, error) {
	var x, y float32
	_, err := fmt.Sscanf(line, "vt %f %f", &x, &y)
	return UV{x, y}, err
}

func parseVertexNormal(line string) (Vec3, error) {
	var x, y, z float32
	_, err := fmt.Sscanf(line, "vn %f %f %f", &x, &y, &z)
	return Vec3{x, y, z}, err
}

func parseFace(c *ObjContext, line string) (Face, error) {
	if strings.Count(line, " ") != 3 {
		return Face{}, errors.New("mesh is not triangulated")
	}

	var (
		err  error
		face Face
	)

	switch {
	case strings.Contains(line, "//"):
		var vn0, vn1, vn2 int

		_, err = fmt.Sscanf(
			line, "f %d//%d %d//%d %d//%d",
			&face.VertexIndices[0], &vn0,
			&face.VertexIndices[1], &vn1,
			&face.VertexIndices[2], &vn2,
		)

		face.VertexNormals[0] = c.VertexNormals[vn0-1]
		face.VertexNormals[1] = c.VertexNormals[vn1-1]
		face.VertexNormals[2] = c.VertexNormals[vn2-1]

	case strings.Contains(line, "/") && strings.Count(line, "/") == 3:
		_, err = fmt.Sscanf(
			line, "f %d/%d %d/%d %d/%d",
			&face.VertexIndices[0], &face.UVs[0],
			&face.VertexIndices[1], &face.UVs[1],
			&face.VertexIndices[2], &face.UVs[2],
		)

	case strings.Contains(line, "/") && strings.Count(line, "/") == 6:
		var vt0, vt1, vt2 int
		var vn0, vn1, vn2 int

		_, err = fmt.Sscanf(
			line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
			&face.VertexIndices[0], &vt0, &vn0,
			&face.VertexIndices[1], &vt1, &vn1,
			&face.VertexIndices[2], &vt2, &vn2,
		)

		face.UVs[0] = c.TextureVertices[vt0-1]
		face.UVs[1] = c.TextureVertices[vt1-1]
		face.UVs[2] = c.TextureVertices[vt2-1]
		face.VertexNormals[0] = c.VertexNormals[vn0-1]
		face.VertexNormals[1] = c.VertexNormals[vn1-1]
		face.VertexNormals[2] = c.VertexNormals[vn2-1]

	default:
		_, err = fmt.Sscanf(
			line, "f %d %d %d",
			&face.VertexIndices[0],
			&face.VertexIndices[1],
			&face.VertexIndices[2],
		)
	}

	// Indices are 1-based in .obj files.
	face.VertexIndices[0] -= 1
	face.VertexIndices[1] -= 1
	face.VertexIndices[2] -= 1

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

						texturePath := m.MapKd
						if texturePath[0] != '/' {
							texturePath = path.Join(dirname, m.MapKd)
						}

						texture, err = LoadTextureFile(texturePath)
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

		case strings.HasPrefix(line, "vn "):
			vn, err := parseVertexNormal(line)
			if err != nil {
				return nil, err
			}
			c.VertexNormals = append(c.VertexNormals, vn)

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
