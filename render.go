package main

import (
	"fmt"
	"image/color"
	"math"
	"sort"
)

var (
	vertexColor = color.RGBA{255, 161, 0, 255}
	faceColor   = color.RGBA{255, 255, 255, 255}
	edgeColor   = color.RGBA{0, 0, 0, 255}
)

type DebugInfo struct {
	X, Y int
	Text string
}

type Renderer struct {
	fb        *FrameBuffer
	triangles []Triangle

	ShowVertices    bool
	ShowEdges       bool
	ShowFaces       bool
	BackfaceCulling bool
	Lighting        bool

	Debug     bool
	DebugInfo []DebugInfo
}

func NewRenderer(fb *FrameBuffer) *Renderer {
	return &Renderer{
		fb:              fb,
		ShowFaces:       true,
		BackfaceCulling: true,
		Lighting:        true,
	}
}

func projectPoint(v Vec4, perspective *Matrix, center Vec2) Vec4 {
	v = perspective.MultiplyVec4(v)

	// Coordinates are normalized in the range [-1, 1]
	if v.W != 0 {
		v.X /= v.W
		v.Y /= v.W
		v.Z /= v.W
	}

	// Scale and translate to screen center.
	// The Y axis is inverted because in screen coordinates the origin is in the top-left corner.
	v.X = v.X*center.X + center.X
	v.Y = v.Y*center.Y*-1 + center.Y

	return v
}

func colorIntensity(c color.RGBA, intensity float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * intensity),
		G: uint8(float64(c.G) * intensity),
		B: uint8(float64(c.B) * intensity),
		A: c.A,
	}
}

func (r *Renderer) project(mesh *Mesh, camera *Camera) {
	r.triangles = r.triangles[:0]
	r.DebugInfo = r.DebugInfo[:0]

	// Calculate the world matrix, which is essentially a series of transformations
	// applied to the mesh vertices that convert them from object's local space to
	// world space. NOTE: The order is important!
	worldMatrix := NewIdentity()
	worldMatrix = NewScale(mesh.Scale.X, mesh.Scale.Y, mesh.Scale.Z).Multiply(worldMatrix)
	worldMatrix = NewRotationZ(mesh.Rotation.Z).Multiply(worldMatrix)
	worldMatrix = NewRotationY(mesh.Rotation.Y).Multiply(worldMatrix)
	worldMatrix = NewRotationX(mesh.Rotation.X).Multiply(worldMatrix)
	worldMatrix = NewTranslation(mesh.Translation.X, mesh.Translation.Y, mesh.Translation.Z).Multiply(worldMatrix)

	// Calculate the parameters for the perspective projection
	aspect := float64(r.fb.Width) / float64(r.fb.Height)
	fovRadians := camera.FOVAngle * (math.Pi / 180.0)
	zFar := 100.0
	zNear := 0.1

	center := Vec2{X: float64(r.fb.Width) / 2, Y: float64(r.fb.Height) / 2}
	perspective := NewPerspective(fovRadians, aspect, zNear, zFar)
	faceColor = color.RGBA{200, 200, 200, 255}

	// Calculate the light vector
	lightVector := Vec3{X: -0.5, Y: 0.5, Z: -1}.Normalize()

	for i, face := range mesh.Faces {
		v1 := worldMatrix.MultiplyVec4(mesh.Vertices[face.A].ToVec4())
		v2 := worldMatrix.MultiplyVec4(mesh.Vertices[face.B].ToVec4())
		v3 := worldMatrix.MultiplyVec4(mesh.Vertices[face.C].ToVec4())

		a3, b3, c3 := v1.ToVec3(), v2.ToVec3(), v3.ToVec3()
		faceNormal := b3.Sub(a3).CrossProduct(c3.Sub(a3)).Normalize()
		faceColor = color.RGBA{255, 255, 255, 255}

		if r.Lighting {
			lightIntensity := faceNormal.DotProduct(lightVector) * 0.95
			faceColor = colorIntensity(faceColor, max(0.05, min(lightIntensity, 1.0)))
		}

		//if r.BackfaceCulling {
		//	cameraRay := camera.Position.Sub(a3)
		//	if faceNormal.DotProduct(cameraRay) < 0 {
		//		continue
		//	}
		//}

		// Project the vertices to the screen coordinates
		s0 := projectPoint(v1, &perspective, center).ToVec3()
		s1 := projectPoint(v2, &perspective, center).ToVec3()
		s2 := projectPoint(v3, &perspective, center).ToVec3()

		if r.BackfaceCulling {
			if (s1.X-s0.X)*(s2.Y-s0.Y)-(s2.X-s0.X)*(s1.Y-s0.Y) < 0 {
				continue
			}
		}

		// Add the triangle to the list of triangles to be rendered
		r.triangles = append(r.triangles, Triangle{
			A:         Vec2{X: s0.X, Y: s0.Y},
			B:         Vec2{X: s1.X, Y: s1.Y},
			C:         Vec2{X: s2.X, Y: s2.Y},
			Z:         (s0.Z + s1.Z + s2.Z) / 3.0,
			Color:     faceColor,
			FaceIndex: i,
		})
	}
}

func (r *Renderer) rasterize() {
	// Sort triangles by Z coordinate (painter's algorithm)
	sort.Slice(r.triangles, func(i, j int) bool {
		return r.triangles[i].Z < r.triangles[j].Z
	})

	// Draw triangles to the frame buffer
	for _, t := range r.triangles {
		a, b, c := t.A, t.B, t.C

		if r.ShowFaces {
			r.fb.Triangle(
				int(a.X), int(a.Y),
				int(b.X), int(b.Y),
				int(c.X), int(c.Y),
				t.Color,
				t.Z,
			)
		}

		if r.ShowEdges {
			ec := edgeColor
			if !r.ShowFaces {
				// Black edges are not visible when faces are not drawn
				ec = color.RGBA{255, 255, 255, 255}
			}

			r.fb.Line(int(a.X), int(a.Y), int(b.X), int(b.Y), ec, t.Z)
			r.fb.Line(int(b.X), int(b.Y), int(c.X), int(c.Y), ec, t.Z)
			r.fb.Line(int(c.X), int(c.Y), int(a.X), int(a.Y), ec, t.Z)

			if r.ShowFaces {
				center := Vec2{X: (a.X + b.X + c.X) / 3, Y: (a.Y + b.Y + c.Y) / 3}
				r.fb.Rect(int(center.X)-1, int(center.Y)-1, 3, 3, ec, t.Z)

				if r.Debug {
					r.DebugInfo = append(r.DebugInfo, DebugInfo{
						X:    int(center.X),
						Y:    int(center.Y) - 5,
						Text: fmt.Sprintf("%d", t.FaceIndex),
					})
				}
			}
		}

		if r.ShowVertices {
			r.fb.Rect(int(a.X)-1, int(a.Y)-1, 3, 3, vertexColor, t.Z)
			r.fb.Rect(int(b.X)-1, int(b.Y)-1, 3, 3, vertexColor, t.Z)
			r.fb.Rect(int(c.X)-1, int(c.Y)-1, 3, 3, vertexColor, t.Z)
		}
	}
}

func (r *Renderer) Draw(mesh *Mesh, camera *Camera) {
	r.fb.Fill(color.RGBA{30, 30, 30, 255})
	r.fb.DotGrid(color.RGBA{100, 100, 100, 255}, 10)
	r.project(mesh, camera)
	r.rasterize()
}
