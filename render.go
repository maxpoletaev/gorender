package main

import (
	"image/color"
	"sort"
)

var (
	vertexColor = color.RGBA{255, 161, 0, 255}
	edgeColor   = color.RGBA{255, 255, 255, 255}
	faceColor   = color.RGBA{150, 150, 150, 255}
)

type Renderer struct {
	fb        *FrameBuffer
	triangles []Triangle

	ShowVertices    bool
	ShowEdges       bool
	ShowFaces       bool
	BackfaceCulling bool
}

func NewRenderer(fb *FrameBuffer) *Renderer {
	return &Renderer{
		fb:              fb,
		ShowEdges:       true,
		ShowFaces:       true,
		BackfaceCulling: true,
	}
}

func project2D(v Vec3) Vec2 {
	return Vec2{
		X: (v.X * fovFactor) / v.Z,
		Y: (v.Y * fovFactor) / v.Z,
	}
}

func (r *Renderer) resetTriangles(n int) {
	if cap(r.triangles) < n {
		r.triangles = make([]Triangle, 0, n)
	} else {
		r.triangles = r.triangles[:0]
	}
}

func (r *Renderer) Draw(mesh *Mesh, camera *Camera) {
	r.fb.Fill(color.RGBA{30, 30, 30, 255})
	r.fb.DotGrid(color.RGBA{100, 100, 100, 255}, 10)
	r.resetTriangles(len(mesh.Faces))

	transforms := NewIdentity().
		Multiply(NewTranslation(mesh.Translation.X, mesh.Translation.Y, mesh.Translation.Z)).
		Multiply(NewScale(mesh.Scale.X, mesh.Scale.Y, mesh.Scale.Z)).
		Multiply(NewRotationX(mesh.Rotation.X)).
		Multiply(NewRotationY(mesh.Rotation.Y)).
		Multiply(NewRotationZ(mesh.Rotation.Z))

	for _, face := range mesh.Faces {
		a := transforms.MultiplyVec4(mesh.Vertices[face.A].ToVec4()).ToVec3()
		b := transforms.MultiplyVec4(mesh.Vertices[face.B].ToVec4()).ToVec3()
		c := transforms.MultiplyVec4(mesh.Vertices[face.C].ToVec4()).ToVec3()

		if r.BackfaceCulling {
			cameraRay := camera.Position.Sub(a)
			faceNormal := b.Sub(a).CrossProduct(c.Sub(a)).Normalize()

			if faceNormal.DotProduct(cameraRay) < 0 {
				continue
			}
		}

		t := Triangle{A: project2D(a), B: project2D(b), C: project2D(c)}
		t.AvgZ = (a.Z + b.Z + c.Z) / 3.0

		r.triangles = append(r.triangles, t)
	}

	// Sort triangles by Z coordinate (painter's algorithm)
	sort.Slice(r.triangles, func(i, j int) bool {
		return r.triangles[i].AvgZ < r.triangles[j].AvgZ
	})

	center := Vec2{
		X: float32(r.fb.Width) / 2.0,
		Y: float32(r.fb.Height) / 2.0,
	}

	for _, t := range r.triangles {
		// Translate to the center
		a := t.A.Add(center)
		b := t.B.Add(center)
		c := t.C.Add(center)

		if r.ShowFaces {
			r.fb.Triangle(
				int(a.X), int(a.Y),
				int(b.X), int(b.Y),
				int(c.X), int(c.Y),
				faceColor,
				t.AvgZ,
			)
		}

		if r.ShowEdges {
			r.fb.Line(int(a.X), int(a.Y), int(b.X), int(b.Y), edgeColor, t.AvgZ)
			r.fb.Line(int(b.X), int(b.Y), int(c.X), int(c.Y), edgeColor, t.AvgZ)
			r.fb.Line(int(c.X), int(c.Y), int(a.X), int(a.Y), edgeColor, t.AvgZ)
		}

		if r.ShowVertices {
			r.fb.Rect(int(a.X)-1, int(a.Y)-1, 3, 3, vertexColor, t.AvgZ)
			r.fb.Rect(int(b.X)-1, int(b.Y)-1, 3, 3, vertexColor, t.AvgZ)
			r.fb.Rect(int(c.X)-1, int(c.Y)-1, 3, 3, vertexColor, t.AvgZ)
		}
	}
}
