package main

import (
	"image/color"
	"math"
	"runtime"
	"sync"
)

const (
	maxTiles = 16
)

var (
	faceColor   = color.RGBA{200, 200, 200, 255}
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
	Points    [3]Vec4
	UVs       [3]UV
	Intensity [3]float32
	Texture   *Texture
	//Intensity   float32
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
	tile uint
}

func calculateTileBoundaries(tile uint, numTiles uint, width, height int) (start, end Vec2) {
	if numTiles == 1 {
		return Vec2{0, 0}, Vec2{float32(width), float32(height)}
	}

	var (
		numTilesX  = uint(math.Sqrt(float64(numTiles)))
		numTilesY  = (numTiles + numTilesX - 1) / numTilesX
		tileWidth  = (uint(width) + numTilesX - 1) / numTilesX
		tileHeight = (uint(height) + numTilesY - 1) / numTilesY
	)

	start.X = float32((tile % numTilesX) * tileWidth)
	start.Y = float32((tile / numTilesX) * tileHeight)
	end.X = start.X + float32(tileWidth)
	end.Y = start.Y + float32(tileHeight)

	if end.X > float32(width) {
		end.X = float32(width)
	}

	if end.Y > float32(height) {
		end.Y = float32(height)
	}

	return start, end
}

type LocalBuffer struct {
	tileTriangles     [maxTiles][128]Triangle
	tileTriangleCount [maxTiles]int
}

type Renderer struct {
	fb               *FrameBuffer
	frustum          *Frustum
	aspectX, aspectY float32
	zNear, zFar      float32
	fovX, fovY       float32

	FrustumClipping bool
	ShowVertices    bool
	ShowEdges       bool
	ShowFaces       bool
	BackfaceCulling bool
	Lighting        bool
	ShowTextures    bool

	DebugEnabled bool
	DebugInfo    []DebugInfo

	toProject chan projectionTask
	toDraw    chan rasterizationTask
	wg        sync.WaitGroup

	numTiles      uint
	tileBounds    [maxTiles][2]Vec2
	tileTriangles [maxTiles][]Triangle
	tileLocks     [maxTiles]sync.Mutex
	localBufPool  *sync.Pool // *LocalBuffer
}

func NewRenderer(fb *FrameBuffer) *Renderer {
	aspectX := float32(fb.Width) / float32(fb.Height)
	aspectY := float32(fb.Height) / float32(fb.Width)

	fovY := float32(45 * (math.Pi / 180))
	fovX := float32(2 * math.Atan(math.Tan(float64(fovY/2))*float64(aspectX)))

	zNear, zFar := float32(0.0), float32(50.0)
	frustum := NewFrustum(zNear, zFar)

	localBufPool := &sync.Pool{
		New: func() interface{} {
			return &LocalBuffer{}
		},
	}

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
		localBufPool:    localBufPool,
	}

	if parallel {
		r.numTiles = max(uint(runtime.NumCPU()), maxTiles)

		for i := uint(0); i < r.numTiles; i++ {
			go r.startWorker(i)
		}
	}

	for i := uint(0); i < r.numTiles; i++ {
		start, end := calculateTileBoundaries(i, r.numTiles, fb.Width, fb.Height)
		r.tileBounds[i] = [2]Vec2{start, end}
	}

	return r
}

func (r *Renderer) drawProjection(t *Triangle, tile uint) {
	lightA, lightB, lightC := t.Intensity[0], t.Intensity[1], t.Intensity[2]
	a, b, c := t.Points[0], t.Points[1], t.Points[2]
	uvA, uvB, uvC := t.UVs[0], t.UVs[1], t.UVs[2]

	var (
		tileStart = r.tileBounds[tile][0]
		tileEnd   = r.tileBounds[tile][1]
	)

	if !r.ShowTextures {
		t.Texture = nil
	}

	if r.ShowFaces {
		r.fb.Triangle(
			int(a.X), int(a.Y), a.W, uvA.U, uvA.V,
			int(b.X), int(b.Y), b.W, uvB.U, uvB.V,
			int(c.X), int(c.Y), c.W, uvC.U, uvC.V,
			int(tileStart.X), int(tileStart.Y), int(tileEnd.X), int(tileEnd.Y),
			lightA, lightB, lightC,
			t.Texture,
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

func (r *Renderer) renderTile(tile uint) {
	for i := range r.tileTriangles[tile] {
		r.drawProjection(&r.tileTriangles[tile][i], tile)
	}
}

// identifyTriangleTiles returns a bitfield of tile numbers that the triangle is visible in.
func (r *Renderer) identifyTriangleTiles(points *[3]Vec4, tileNums *[maxTiles]uint8) (n int) {
	var (
		// Triangle bounding box
		minX = min(points[0].X, points[1].X, points[2].X)
		maxX = max(points[0].X, points[1].X, points[2].X)
		minY = min(points[0].Y, points[1].Y, points[2].Y)
		maxY = max(points[0].Y, points[1].Y, points[2].Y)
	)

	for i := uint(0); i < r.numTiles; i++ {
		start, end := &r.tileBounds[i][0], &r.tileBounds[i][1]
		if maxX >= start.X && minX <= end.X && maxY >= start.Y && minY <= end.Y {
			tileNums[n] = uint8(i)
			n++
		}
	}

	return n
}

// projectObject projects the object to the screen space. Object’s Face projections are
// stored in the corresponding tileTriangle buffers for later rasterization.
func (r *Renderer) projectObject(object *Object, camera *Camera) {
	worldMatrix := NewWorldMatrix(object.Scale, object.Rotation, object.Translation)
	viewMatrix := NewViewMatrix(camera.Position, camera.Direction, camera.Up)
	perspectiveMatrix := NewPerspectiveMatrix(r.fovY, r.aspectX, r.zNear, r.zFar)

	mvpMatrix := NewIdentityMatrix()
	mvpMatrix = mvpMatrix.Multiply(perspectiveMatrix)
	mvpMatrix = mvpMatrix.Multiply(viewMatrix)
	mvpMatrix = mvpMatrix.Multiply(worldMatrix)

	screenMatrix := NewScreenMatrix(r.fb.Width, r.fb.Height)
	lightDirection := Vec3{X: -0.5, Y: 1, Z: 1}.Normalize()

	// Transform the bounding box to clip space
	bbox := object.BoundingBox
	matrixMultiplyVec4Batch(&mvpMatrix, bbox[:])

	// Quick check if the object is inside the frustum
	boxVisibility := r.frustum.BoxVisibility(&bbox)
	if boxVisibility == BoxVisibilityOutside {
		return
	}

	var (
		tileNums [maxTiles]uint8

		// Original triangle points
		points          [3]Vec4
		vertexNormals   [3]Vec3
		vertexIntensity [3]float32

		// New points after frustum clipping
		clipIntensity [maxClipPoints][3]float32
		clipPoints    [maxClipPoints][3]Vec4
		clipUV        [maxClipPoints][3]UV
		clipCount     int
	)

	// Local buffers are pooled to avoid zeroing them on each frame
	localBuf := r.localBufPool.Get().(*LocalBuffer)
	defer r.localBufPool.Put(localBuf)
	tileTriangles := &localBuf.tileTriangles
	tileTriangleCount := &localBuf.tileTriangleCount

	// Reset the counts for each tile just in case.
	for i := range tileTriangleCount {
		tileTriangleCount[i] = 0
	}

	// Transform the vertices to clip space
	copy(object.Transformed, object.Vertices)
	matrixMultiplyVec4Batch(&mvpMatrix, object.Transformed)

	for fi := range object.Faces {
		face := &object.Faces[fi] // avoid face copy

		points[0] = object.Transformed[face.VertexIndices[0]]
		points[1] = object.Transformed[face.VertexIndices[1]]
		points[2] = object.Transformed[face.VertexIndices[2]]

		// Transform the vertex normals to world space
		for i := range face.VertexNormals {
			vn := face.VertexNormals[i].ToVec4()
			matrixMultiplyVec4Inplace(&worldMatrix, &vn)
			vertexNormals[i] = vn.ToVec3().Normalize()
		}

		// Calculate the face normal (cross product of two edges)
		v0, v1, v2 := points[0].ToVec3(), points[1].ToVec3(), points[2].ToVec3()
		faceNormal := v1.Sub(v0).CrossProduct(v2.Sub(v0))

		if r.BackfaceCulling {
			// Normal-based backface culling (face is not visible if the normal is not facing the camera)
			if faceNormal.DotProduct(Vec3{0, 0, 0}.Sub(v0)) < 0 {
				continue
			}
		}

		if r.Lighting {
			for i := range vertexNormals {
				diffuse := max(vertexNormals[i].DotProduct(lightDirection), 0.0) * 0.5
				vertexIntensity[i] = 0.5 + diffuse
			}
		}

		// Clip triangles if object is not fully inside the frustum
		if r.FrustumClipping && boxVisibility != BoxVisibilityInside {
			clipCount = r.frustum.ClipTriangle(
				&points, &face.UVs, &vertexIntensity,
				&clipPoints, &clipUV, &clipIntensity,
			)
		} else {
			clipIntensity[0] = vertexIntensity
			clipPoints[0] = points
			clipUV[0] = face.UVs
			clipCount = 1
		}

		for i := 0; i < clipCount; i++ {
			screenPoints := clipPoints[i]

			// Perspective divide
			for j := range screenPoints {
				origW := screenPoints[j].W
				screenPoints[j] = screenPoints[j].Divide(screenPoints[j].W)
				matrixMultiplyVec4Inplace(&screenMatrix, &screenPoints[j])
				screenPoints[j].W = origW
			}

			triangle := Triangle{
				Points:    screenPoints,
				UVs:       clipUV[i],
				Texture:   face.Texture,
				Intensity: clipIntensity[i],
			}

			// Identify the tiles that the triangle is visible in
			for n := range r.identifyTriangleTiles(&screenPoints, &tileNums) {
				tile := tileNums[n]

				// Add triangle to the corresponding local tile buffer
				tileTriangles[tile][tileTriangleCount[tile]] = triangle
				tileTriangleCount[tile]++

				// Flush local buffer to the global buffer once it's full
				if tileTriangleCount[tile] == len(tileTriangles[tile]) {
					r.tileLocks[tile].Lock()
					r.tileTriangles[tile] = append(r.tileTriangles[tile], tileTriangles[tile][:]...)
					r.tileLocks[tile].Unlock()
					tileTriangleCount[tile] = 0
				}
			}
		}
	}

	// Flush the remaining triangles to the global buffer
	for tile := range tileTriangles {
		if tileTriangleCount[tile] != 0 {
			r.tileLocks[tile].Lock()
			r.tileTriangles[tile] = append(r.tileTriangles[tile], tileTriangles[tile][:tileTriangleCount[tile]]...)
			r.tileLocks[tile].Unlock()
			tileTriangleCount[tile] = 0
		}
	}
}

func (r *Renderer) startWorker(tile uint) {
	for {
		select {
		case task := <-r.toProject:
			r.projectObject(task.object, task.camera)
			r.wg.Done()
		case <-r.toDraw:
			r.renderTile(tile)
			r.wg.Done()
		}
	}
}

func (r *Renderer) drawTilesBoundaries() {
	for i := uint(0); i < r.numTiles; i++ {
		var (
			start = r.tileBounds[i][0]
			end   = r.tileBounds[i][1]
		)

		r.fb.Line(int(start.X), int(start.Y), int(end.X), int(start.Y), color.RGBA{255, 0, 0, 255})
		r.fb.Line(int(end.X), int(start.Y), int(end.X), int(end.Y), color.RGBA{255, 0, 0, 255})
		r.fb.Line(int(end.X), int(end.Y), int(start.X), int(end.Y), color.RGBA{255, 0, 0, 255})
		r.fb.Line(int(start.X), int(end.Y), int(start.X), int(start.Y), color.RGBA{255, 0, 0, 255})
	}
}

func (r *Renderer) Draw(objects []*Object, camera *Camera) {
	for i := uint(0); i < r.numTiles; i++ {
		r.tileTriangles[i] = r.tileTriangles[i][:0]
	}

	r.fb.Clear(color.RGBA{50, 50, 50, 255})
	r.fb.DotGrid(color.RGBA{100, 100, 100, 255}, 10)

	if parallel {
		r.wg.Add(len(objects))
		for i := range objects {
			r.toProject <- projectionTask{
				object: objects[i],
				camera: camera,
			}
		}
		r.wg.Wait()

		r.wg.Add(int(r.numTiles))
		for i := uint(0); i < r.numTiles; i++ {
			r.toDraw <- rasterizationTask{tile: i}
		}
		r.wg.Wait()
	} else {
		for i := range objects {
			r.projectObject(objects[i], camera)
		}

		for i := uint(0); i < r.numTiles; i++ {
			r.renderTile(i)
		}
	}

	if !demoMode {
		//r.drawTilesBoundaries()
		r.fb.CrossHair(color.RGBA{255, 255, 0, 255})
		//r.fb.Fog(0.100, 0.033, color.RGBA{100, 100, 100, 255})
	}
}
