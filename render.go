package main

import (
	"image/color"
	"math"
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

func projectPoint(v Vec4, perspective *Matrix, center Vec2) Vec4 {
	v = perspective.MultiplyVec4(v)

	if v.W != 0 {
		v.X /= v.W
		v.Y /= v.W
		v.Z /= v.W
	}

	// Scale and translate to screen center
	v.X = v.X*center.X + center.X
	v.Y = v.Y*center.Y + center.Y

	return v
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

	// World matrix is essentially a series of transformations applied to the mesh
	// vertices that convert them from object's local space (model space) to world
	// space. NOTE: The order is important!
	worldMatrix := NewIdentity()
	worldMatrix = NewScale(mesh.Scale.X, mesh.Scale.Y, mesh.Scale.Z).Multiply(worldMatrix)
	worldMatrix = NewRotationZ(mesh.Rotation.Z).Multiply(worldMatrix)
	worldMatrix = NewRotationY(mesh.Rotation.Y).Multiply(worldMatrix)
	worldMatrix = NewRotationX(mesh.Rotation.X).Multiply(worldMatrix)
	worldMatrix = NewTranslation(mesh.Translation.X, mesh.Translation.Y, mesh.Translation.Z).Multiply(worldMatrix)

	// Calculate the parameters for the perspective projection
	aspect := float64(r.fb.Width) / float64(r.fb.Height)
	fovRadians := camera.FOVAngle * (math.Pi / 180.0)
	zFar := -100.0
	zNear := 0.1

	center := Vec2{X: float64(r.fb.Width) / 2, Y: float64(r.fb.Height) / 2}
	perspective := NewPerspective(fovRadians, aspect, zNear, zFar)

	for _, face := range mesh.Faces {
		a := worldMatrix.MultiplyVec4(mesh.Vertices[face.A].ToVec4())
		b := worldMatrix.MultiplyVec4(mesh.Vertices[face.B].ToVec4())
		c := worldMatrix.MultiplyVec4(mesh.Vertices[face.C].ToVec4())

		if r.BackfaceCulling {
			a3, b3, c3 := a.ToVec3(), b.ToVec3(), c.ToVec3()
			faceNormal := b3.Sub(a3).CrossProduct(c3.Sub(a3))
			cameraRay := camera.Position.Sub(a3)

			if faceNormal.DotProduct(cameraRay) < 0 {
				continue
			}
		}

		// Project the vertices to the screen coordinates
		pa := projectPoint(a, &perspective, center)
		pb := projectPoint(b, &perspective, center)
		pc := projectPoint(c, &perspective, center)

		// Add the triangle to the list of triangles to be rendered
		r.triangles = append(r.triangles, Triangle{
			A: Vec2{X: pa.X, Y: pa.Y},
			B: Vec2{X: pb.X, Y: pb.Y},
			C: Vec2{X: pc.X, Y: pc.Y},
			Z: (a.Z + b.Z + c.Z) / 3.0,
		})
	}

	// Sort triangles by Z coordinate (painter's algorithm)
	sort.Slice(r.triangles, func(i, j int) bool {
		return r.triangles[i].Z < r.triangles[j].Z
	})

	// Rasterize the triangles to the frame buffer
	for _, t := range r.triangles {
		a, b, c := t.A, t.B, t.C

		if r.ShowFaces {
			r.fb.Triangle(
				int(a.X), int(a.Y),
				int(b.X), int(b.Y),
				int(c.X), int(c.Y),
				faceColor,
				t.Z,
			)
		}

		if r.ShowEdges {
			r.fb.Line(int(a.X), int(a.Y), int(b.X), int(b.Y), edgeColor, t.Z)
			r.fb.Line(int(b.X), int(b.Y), int(c.X), int(c.Y), edgeColor, t.Z)
			r.fb.Line(int(c.X), int(c.Y), int(a.X), int(a.Y), edgeColor, t.Z)
		}

		if r.ShowVertices {
			r.fb.Rect(int(a.X)-1, int(a.Y)-1, 3, 3, vertexColor, t.Z)
			r.fb.Rect(int(b.X)-1, int(b.Y)-1, 3, 3, vertexColor, t.Z)
			r.fb.Rect(int(c.X)-1, int(c.Y)-1, 3, 3, vertexColor, t.Z)
		}
	}
}
