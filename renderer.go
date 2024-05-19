package main

import (
	"image/color"
	"math"
	"runtime"
	"sync"
)

const (
	parallelRendering = true
)

var (
	vertexColor = color.RGBA{255, 161, 0, 255}
	edgeColor   = color.RGBA{0, 0, 0, 255}
)

type Camera struct {
	Position  Vec3
	Direction Vec3
	Up        Vec3
}

// Projection is a 2D projection of a Face.
type Projection struct {
	Points    [3]Vec4
	UVs       [3]UV
	Texture   *Texture
	Intensity float64
}

type DebugInfo struct {
	X, Y int
	Text string
}

type RenderTask struct {
	Object *Object
	Camera *Camera
}

type Renderer struct {
	fb               *FrameBuffer
	frustum          *Frustum
	aspectX, aspectY float64
	zNear, zFar      float64
	fovX, fovY       float64

	wg    sync.WaitGroup
	tasks chan RenderTask

	FrustumClipping bool
	ShowVertices    bool
	ShowEdges       bool
	ShowFaces       bool
	BackfaceCulling bool
	Lighting        bool
	ShowTextures    bool

	DebugEnabled bool
	DebugInfo    []DebugInfo
}

func NewRenderer(fb *FrameBuffer) *Renderer {
	aspectX := float64(fb.Width) / float64(fb.Height)
	aspectY := float64(fb.Height) / float64(fb.Width)

	fovY := 45 * (math.Pi / 180)
	fovX := 2 * math.Atan(math.Tan(fovY/2)*aspectX)

	zNear, zFar := 0.0, 100.0
	frustum := NewFrustum(zNear, zFar)

	r := &Renderer{
		fb:              fb,
		ShowFaces:       true,
		BackfaceCulling: true,
		Lighting:        true,
		FrustumClipping: true,
		ShowTextures:    true,
		fovX:            fovX,
		fovY:            fovY,
		aspectX:         aspectX,
		aspectY:         aspectY,
		frustum:         frustum,
		zNear:           zNear,
		zFar:            zFar,
		tasks:           make(chan RenderTask),
	}

	if parallelRendering {
		for i := 0; i < runtime.NumCPU(); i++ {
			go r.startWorker()
		}
	}

	return r
}

func (r *Renderer) drawProjection(t *Projection) {
	a, b, c := t.Points[0], t.Points[1], t.Points[2]
	uvA, uvB, uvC := t.UVs[0], t.UVs[1], t.UVs[2]

	center := Vec2{
		X: (a.X + b.X + c.X) / 3,
		Y: (a.Y + b.Y + c.Y) / 3,
	}

	if !r.ShowTextures {
		t.Texture = nil
	}

	if r.ShowFaces {
		r.fb.Triangle2(
			int(a.X), int(a.Y), a.W, uvA.U, uvA.V,
			int(b.X), int(b.Y), b.W, uvB.U, uvB.V,
			int(c.X), int(c.Y), c.W, uvC.U, uvC.V,
			t.Texture,
			t.Intensity,
		)
	}

	if r.ShowEdges {
		ec := edgeColor
		if !r.ShowFaces {
			// Black edges are not visible when faces are not drawn
			ec = color.RGBA{255, 255, 255, 255}
		}

		r.fb.Line(int(a.X), int(a.Y), int(b.X), int(b.Y), ec)
		r.fb.Line(int(b.X), int(b.Y), int(c.X), int(c.Y), ec)
		r.fb.Line(int(c.X), int(c.Y), int(a.X), int(a.Y), ec)

		if r.ShowFaces {
			r.fb.Rect(int(center.X)-1, int(center.Y)-1, 3, 3, ec)
		}
	}

	if r.ShowVertices {
		r.fb.Rect(int(a.X)-1, int(a.Y)-1, 3, 3, vertexColor)
		r.fb.Rect(int(b.X)-1, int(b.Y)-1, 3, 3, vertexColor)
		r.fb.Rect(int(c.X)-1, int(c.Y)-1, 3, 3, vertexColor)
	}
}

func (r *Renderer) renderObject(object *Object, camera *Camera) {
	worldMatrix := NewWorldMatrix(object.Scale, object.Rotation, object.Translation)
	viewMatrix := NewViewMatrix(camera.Position, camera.Direction, camera.Up)
	perspectiveMatrix := NewPerspectiveMatrix(r.fovY, r.aspectX, r.zNear, r.zFar)
	screenMatrix := NewScreenMatrix(r.fb.Width, r.fb.Height)
	lightDirection := Vec3{Y: 0.5, Z: -1}.Normalize()
	bbox := object.BoundingBox

	// Apply transformations to the bounding box
	for i := 0; i < 8; i++ {
		bbox[i] = worldMatrix.MultiplyVec4(bbox[i])
		bbox[i] = viewMatrix.MultiplyVec4(bbox[i])
		bbox[i] = perspectiveMatrix.MultiplyVec4(bbox[i])
	}

	// Quick check if the object is inside the frustum
	if !r.frustum.IsBoxVisible(bbox) {
		return
	}

	for fi := range object.Faces {
		face := &object.Faces[fi]

		points := [3]Vec4{
			object.Vertices[face.A].ToVec4(),
			object.Vertices[face.B].ToVec4(),
			object.Vertices[face.C].ToVec4(),
		}

		for p := range points {
			points[p] = worldMatrix.MultiplyVec4(points[p])       // Local -> World space
			points[p] = viewMatrix.MultiplyVec4(points[p])        // World -> View space
			points[p] = perspectiveMatrix.MultiplyVec4(points[p]) // View -> Clip space
		}

		a, b, c := points[0].ToVec3(), points[1].ToVec3(), points[2].ToVec3()
		faceNormal := b.Sub(a).CrossProduct(c.Sub(a)).Normalize()

		if r.BackfaceCulling {
			// Normal-based backface culling (face is not visible if the normal is not facing the camera)
			if faceNormal.DotProduct(Vec3{0, 0, 0}.Sub(a).Normalize()) < 0 {
				continue
			}
		}

		var (
			lightIntensity = 0.5
		)

		if r.Lighting {
			const (
				ambientStrength = 0.5
				diffuseStrength = 0.5
			)

			diffuse := math.Max(faceNormal.DotProduct(lightDirection), 0.0) * diffuseStrength
			lightIntensity = ambientStrength + diffuse
		}

		var (
			// Clipping will produce of up to 9 new triangles
			clipPoints [maxClipPoints][3]Vec4
			clipUVs    [maxClipPoints][3]UV
			clipCount  int
		)

		// TODO: check if clipping is needed (all points are inside)
		if r.FrustumClipping {
			uvs := [3]UV{face.UVa, face.UVb, face.UVc}
			clipCount = r.frustum.ClipTriangle(points, uvs, &clipPoints, &clipUVs)
		} else {
			clipPoints[0] = [3]Vec4{points[0], points[1], points[2]}
			clipUVs[0] = [3]UV{face.UVa, face.UVb, face.UVc}
			clipCount = 1
		}

		for i := 0; i < clipCount; i++ {
			newPoints := clipPoints[i]

			for j, v := range newPoints {
				origW := v.W
				v = v.Divide(v.W) // Perspective divide
				v = screenMatrix.MultiplyVec4(Vec4{v.X, v.Y, v.Z, 1})
				v.W = origW // Need the original W for texture mapping
				newPoints[j] = v
			}

			r.drawProjection(&Projection{
				Points:    newPoints,
				Texture:   object.Texture,
				Intensity: lightIntensity,
				UVs:       clipUVs[i],
			})
		}
	}
}

func (r *Renderer) startWorker() {
	for task := range r.tasks {
		r.renderObject(task.Object, task.Camera)
		r.wg.Done()
	}
}

func (r *Renderer) Draw(objects []*Object, camera *Camera) {
	r.fb.Clear(color.RGBA{50, 50, 50, 255})
	r.fb.DotGrid(color.RGBA{100, 100, 100, 255}, 10)

	if parallelRendering {
		r.wg.Add(len(objects))

		for i := range objects {
			r.tasks <- RenderTask{
				Object: objects[i],
				Camera: camera,
			}
		}

		r.wg.Wait()
	} else {
		for i := range objects {
			r.renderObject(objects[i], camera)
		}
	}

	r.fb.CrossHair(color.RGBA{255, 255, 0, 255})
	//r.fb.Fog(0.100, 0.033, color.RGBA{100, 100, 100, 255})
}
