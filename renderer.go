package main

import (
	"image/color"
	"math"
	"runtime"
	"sync"
)

const (
	parallelRendering = true
	maxTiles          = 16
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

// Triangle is a 2D projection of a Face.
type Triangle struct {
	Points      [3]Vec4
	UVs         [3]UV
	Texture     *Texture
	Intensity   float64
	TileNumbers uint16
}

type DebugInfo struct {
	X, Y int
	Text string
}

type projectionTask struct {
	object *Object
	camera *Camera
}

type rasterizationTask struct {
	tile int
}

func calculateTileBoundaries(tile int, numTiles int, width, height int) (start, end Vec2) {
	if numTiles == 1 {
		return Vec2{0, 0}, Vec2{float64(width), float64(height)}
	}

	var (
		numTilesX  = int(math.Sqrt(float64(numTiles)))
		numTilesY  = (numTiles + numTilesX - 1) / numTilesX
		tileWidth  = (width + numTilesX - 1) / numTilesX
		tileHeight = (height + numTilesY - 1) / numTilesY
	)

	start.X = float64((tile % numTilesX) * tileWidth)
	start.Y = float64((tile / numTilesX) * tileHeight)
	end.X = start.X + float64(tileWidth)
	end.Y = start.Y + float64(tileHeight)

	if end.X > float64(width) {
		end.X = float64(width)
	}

	if end.Y > float64(height) {
		end.Y = float64(height)
	}

	return start, end
}

type Renderer struct {
	fb               *FrameBuffer
	frustum          *Frustum
	aspectX, aspectY float64
	zNear, zFar      float64
	fovX, fovY       float64

	FrustumClipping bool
	ShowVertices    bool
	ShowEdges       bool
	ShowFaces       bool
	BackfaceCulling bool
	Lighting        bool
	ShowTextures    bool

	DebugEnabled bool
	DebugInfo    []DebugInfo

	// Parallel rendering stuff
	toProject      chan projectionTask
	toDraw         chan rasterizationTask
	triangles      []Triangle
	tileBoundaries [maxTiles][2]Vec2
	wg             sync.WaitGroup
	mut            sync.Mutex
	numTiles       int
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
		numTiles:        1,
		toProject:       make(chan projectionTask, 256),
		toDraw:          make(chan rasterizationTask, maxTiles),
	}

	if parallelRendering {
		r.numTiles = max(runtime.NumCPU(), maxTiles)

		for i := 0; i < r.numTiles; i++ {
			go r.startWorker()
		}
	}

	for i := 0; i < r.numTiles; i++ {
		start, end := calculateTileBoundaries(i, r.numTiles, fb.Width, fb.Height)
		r.tileBoundaries[i] = [2]Vec2{start, end}
	}

	return r
}

func (r *Renderer) drawProjection(t *Triangle) {
	a, b, c := t.Points[0], t.Points[1], t.Points[2]
	uvA, uvB, uvC := t.UVs[0], t.UVs[1], t.UVs[2]

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
		colr := edgeColor
		if !r.ShowFaces {
			// Black edges are not visible when faces are not drawn
			colr = color.RGBA{255, 255, 255, 255}
		}

		r.fb.Line(int(a.X), int(a.Y), int(b.X), int(b.Y), colr)
		r.fb.Line(int(b.X), int(b.Y), int(c.X), int(c.Y), colr)
		r.fb.Line(int(c.X), int(c.Y), int(a.X), int(a.Y), colr)

		if r.ShowFaces {
			center := Vec2{
				X: (a.X + b.X + c.X) / 3,
				Y: (a.Y + b.Y + c.Y) / 3,
			}

			r.fb.Rect(int(center.X)-1, int(center.Y)-1, 3, 3, colr)
		}
	}

	if r.ShowVertices {
		r.fb.Rect(int(a.X)-1, int(a.Y)-1, 3, 3, vertexColor)
		r.fb.Rect(int(b.X)-1, int(b.Y)-1, 3, 3, vertexColor)
		r.fb.Rect(int(c.X)-1, int(c.Y)-1, 3, 3, vertexColor)
	}
}

func (r *Renderer) renderTile(tileID int) {
	for i := range r.triangles {
		proj := &r.triangles[i]

		if proj.TileNumbers&(1<<tileID) != 0 {
			r.drawProjection(proj)
		}
	}
}

func (r *Renderer) IsTriangleCrossingTile(points *[3]Vec4, tile int) bool {
	var (
		start = r.tileBoundaries[tile][0]
		end   = r.tileBoundaries[tile][1]
	)

	for i := range points {
		if points[i].X >= start.X &&
			points[i].X <= end.X &&
			points[i].Y >= start.Y &&
			points[i].Y <= end.Y {
			return true
		}
	}

	return false
}

func (r *Renderer) projectObject(object *Object, camera *Camera) {
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

	var (
		localBuf     [16]Triangle
		localBufSize int
	)

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
			// Clipping will produce of up to 9 new tileTriangles
			clipPoints [maxClipPoints][3]Vec4
			clipUVs    [maxClipPoints][3]UV
			clipCount  int
		)

		// TODO: check if clipping is needed (all points are inside)
		if r.FrustumClipping {
			uvs := [3]UV{face.UVa, face.UVb, face.UVc}
			clipCount = r.frustum.ClipTriangle(&points, &uvs, &clipPoints, &clipUVs)
		} else {
			clipPoints[0] = [3]Vec4{points[0], points[1], points[2]}
			clipUVs[0] = [3]UV{face.UVa, face.UVb, face.UVc}
			clipCount = 1
		}

		for i := 0; i < clipCount; i++ {
			newPoints := &clipPoints[i]

			for j, v := range newPoints {
				origW := v.W
				v = v.Divide(v.W) // Perspective divide
				v = screenMatrix.MultiplyVec4(Vec4{v.X, v.Y, v.Z, 1})
				v.W = origW // Need the original W for texture mapping
				newPoints[j] = v
			}

			// Keep the triangle in the local buffer to avoid tacking
			// the global mutex too often
			localBuf[localBufSize] = Triangle{
				Points:    *newPoints,
				UVs:       clipUVs[i],
				Texture:   object.Texture,
				Intensity: lightIntensity,
			}

			for t := 0; t < r.numTiles; t++ {
				// Precompute which tiles the triangle should be rendered to
				if r.IsTriangleCrossingTile(newPoints, t) {
					localBuf[localBufSize].TileNumbers |= 1 << t
				}
			}

			// Flush the buffer once it's full
			if localBufSize++; localBufSize == len(localBuf) {
				r.mut.Lock()
				r.triangles = append(r.triangles, localBuf[:]...)
				r.mut.Unlock()
				localBufSize = 0
			}
		}
	}

	// Flush the remaining tileTriangles
	if localBufSize != 0 {
		r.mut.Lock()
		r.triangles = append(r.triangles, localBuf[:]...)
		r.mut.Unlock()
		localBufSize = 0
	}
}

func (r *Renderer) startWorker() {
	for {
		select {
		case task := <-r.toProject:
			r.projectObject(task.object, task.camera)
			r.wg.Done()

		case task := <-r.toDraw:
			r.renderTile(task.tile)
			r.wg.Done()
		}
	}
}

func (r *Renderer) drawTilesBoundaries() {
	for i := 0; i < r.numTiles; i++ {
		var (
			start = r.tileBoundaries[i][0]
			end   = r.tileBoundaries[i][1]
		)

		r.fb.Line(int(start.X), int(start.Y), int(end.X), int(start.Y), color.RGBA{255, 0, 0, 255})
		r.fb.Line(int(end.X), int(start.Y), int(end.X), int(end.Y), color.RGBA{255, 0, 0, 255})
		r.fb.Line(int(end.X), int(end.Y), int(start.X), int(end.Y), color.RGBA{255, 0, 0, 255})
		r.fb.Line(int(start.X), int(end.Y), int(start.X), int(start.Y), color.RGBA{255, 0, 0, 255})
	}
}

func (r *Renderer) Draw(objects []*Object, camera *Camera) {
	r.triangles = r.triangles[:0]
	r.fb.Clear(color.RGBA{50, 50, 50, 255})
	r.fb.DotGrid(color.RGBA{100, 100, 100, 255}, 10)

	if parallelRendering {
		r.wg.Add(len(objects))
		for i := range objects {
			r.toProject <- projectionTask{
				object: objects[i],
				camera: camera,
			}
		}
		r.wg.Wait()

		r.wg.Add(r.numTiles)
		for i := 0; i < r.numTiles; i++ {
			r.toDraw <- rasterizationTask{tile: i}
		}
		r.wg.Wait()
	} else {
		for i := range objects {
			r.projectObject(objects[i], camera)
		}
		for i := 0; i < r.numTiles; i++ {
			r.renderTile(i)
		}
	}

	//r.drawTilesBoundaries()
	r.fb.CrossHair(color.RGBA{255, 255, 0, 255})
	//r.fb.Fog(0.100, 0.033, color.RGBA{100, 100, 100, 255})
}
