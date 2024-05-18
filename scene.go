package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"path"
)

type SceneMeshData struct {
	ID           string  `json:"id"`
	ObjFile      string  `json:"objFile"`
	Texture      string  `json:"texture"`
	TextureScale float64 `json:"textureScale"`
}

type SceneObjectData struct {
	MeshID   string     `json:"meshID"`
	Position [3]float64 `json:"position"`
	Rotation [3]float64 `json:"rotation"`
	Scale    [3]float64 `json:"scale"`
}

type SceneData struct {
	Name    string            `json:"name"`
	Meshes  []SceneMeshData   `json:"meshes"`
	Objects []SceneObjectData `json:"objects"`
}

type Scene struct {
	Objects []*Object
}

func (s *Scene) NumObjects() int {
	return len(s.Objects)
}

func (s *Scene) NumVertices() (n int) {
	for _, obj := range s.Objects {
		n += len(obj.Mesh.Vertices)
	}

	return n
}

func (s *Scene) NumTriangles() (n int) {
	for _, obj := range s.Objects {
		n += len(obj.Mesh.Faces)
	}

	return n
}

func LoadSceneFile(filename string) (*Scene, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open level file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	decoder := json.NewDecoder(f)
	sceneData := SceneData{}

	if err := decoder.Decode(&sceneData); err != nil {
		return nil, fmt.Errorf("failed to read scene manifest: %w", err)
	}

	defaultColor := color.RGBA{200, 200, 200, 255}
	defaultTexture := NewColorTexture(defaultColor)

	var (
		rootDir = path.Dir(filename)
		meshes  = make(map[string]*Mesh)
		objects []*Object
	)

	for _, meshData := range sceneData.Meshes {
		fmt.Println(path.Join(rootDir, meshData.ObjFile))
		mesh, err := LoadMeshFile(path.Join(rootDir, meshData.ObjFile))
		if err != nil {
			return nil, fmt.Errorf("failed to load mesh '%s': %w", meshData.ID, err)
		}

		if meshData.Texture != "" {
			texture, err := LoadTextureFile(path.Join(rootDir, meshData.Texture))
			if err != nil {
				return nil, fmt.Errorf("failed to load texture %s: %w", meshData.ID, err)
			}

			if meshData.TextureScale != 0 {
				texture.SetScale(meshData.TextureScale)
			}

			mesh.Texture = texture
		} else {
			mesh.Texture = defaultTexture
		}

		meshes[meshData.ID] = mesh
	}

	for _, objData := range sceneData.Objects {
		mesh, ok := meshes[objData.MeshID]
		if !ok {
			return nil, fmt.Errorf("mesh id not found: %s", objData.MeshID)
		}

		if objData.Scale == [3]float64{0, 0, 0} {
			log.Printf("[WARN] object scale is zero: %s", objData.MeshID)
		}

		obj := NewObject(mesh)
		obj.Scale = Vec3FromArray(objData.Scale)
		obj.Rotation = Vec3FromArray(objData.Rotation).ToRadians()
		obj.Translation = Vec3FromArray(objData.Position)

		objects = append(objects, obj)
	}

	return &Scene{
		Objects: objects,
	}, err
}
