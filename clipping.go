package main

import "sync"

const (
	// maxClipPoints is the maximum number of vertices a polygon can have.
	// 9 is the worst case scenario for a triangle clipped against all 6 planes.
	maxClipPoints = 9
)

const (
	PlaneLeft = iota
	PlaneRight
	PlaneTop
	PlaneBottom
	PlaneNear
	PlaneFar
)

const (
	BoxVisibilityOutside = iota
	BoxVisibilityIntersect
	BoxVisibilityInside
)

type Polygon struct {
	Points [maxClipPoints]Vec4
	UVs    [maxClipPoints]UV
	Count  int
}

func (p *Polygon) AddVertex(v Vec4, uv UV) {
	p.Points[p.Count] = v
	p.UVs[p.Count] = uv
	p.Count++
}

func (p *Polygon) Triangulate(
	points *[maxClipPoints][3]Vec4,
	uvs *[maxClipPoints][3]UV,
) (numOut int) {
	if p.Count < 3 {
		return 0
	}

	// The resulting polygon is convex, so we can always triangulate it by connecting
	// the first vertex with the second and the third, then the first vertex with the
	// third and the fourth, and so on (fan triangulation).
	for i := 0; i < p.Count-2; i++ {
		points[numOut] = [3]Vec4{p.Points[0], p.Points[i+1], p.Points[i+2]}
		uvs[numOut] = [3]UV{p.UVs[0], p.UVs[i+1], p.UVs[i+2]}
		numOut++
	}

	return numOut
}

type Plane struct {
	Point  Vec4
	Normal Vec4
}

func (p *Plane) DistanceToVertex(v Vec4) float64 {
	return p.Normal.DotProduct(v) - p.Normal.DotProduct(p.Point)
}

// IsVertexInside tells if a point is inside or outside the plane.
func (p *Plane) IsVertexInside(q Vec4) bool {
	return q.Sub(p.Point).DotProduct(p.Normal) <= 0
}

// Intersect returns a point between q0 and q1 intersect with the plane.
func (p *Plane) Intersect(q0, q1 Vec4) (Vec4, float64) {
	u := q1.Sub(q0)
	w := q0.Sub(p.Point)
	d := p.Normal.DotProduct(u)
	n := -p.Normal.DotProduct(w)
	factor := n / d // interpolation factor
	return q0.Add(u.Multiply(factor)), factor
}

type Frustum struct {
	Planes      [6]Plane
	polygonPool *sync.Pool // *Polygon
}

func NewFrustum(zNear, zFar float64) *Frustum {
	polygonPool := &sync.Pool{
		New: func() any {
			return &Polygon{}
		},
	}

	return &Frustum{
		Planes: [6]Plane{
			PlaneLeft: {
				Point:  Vec4{-1, 0, 0, 1},
				Normal: Vec4{1, 0, 0, 1},
			},
			PlaneRight: {
				Point:  Vec4{1, 0, 0, 1},
				Normal: Vec4{-1, 0, 0, 1},
			},
			PlaneTop: {
				Point:  Vec4{0, -1, 0, 1},
				Normal: Vec4{0, 1, 0, 1},
			},
			PlaneBottom: {
				Point:  Vec4{0, 1, 0, 1},
				Normal: Vec4{0, -1, 0, 1},
			},
			PlaneNear: {
				Point:  Vec4{0, 0, zNear, 1},
				Normal: Vec4{0, 0, -1, 0},
			},
			PlaneFar: {
				Point:  Vec4{0, 0, zFar, 1},
				Normal: Vec4{0, 0, 1, 0},
			},
		},
		polygonPool: polygonPool,
	}
}

func (f *Frustum) BoxVisibility(bbox *[8]Vec4) int {
	for i := range f.Planes {
		outside := 0

		for _, point := range bbox {
			if f.Planes[i].DistanceToVertex(point) > 0 {
				outside++
			}
		}

		// All points are outside, the box is not visible
		if outside == len(bbox) {
			return BoxVisibilityOutside
		}

		// Some points are outside, clipping is needed
		if outside > 0 {
			return BoxVisibilityIntersect
		}

	}

	return BoxVisibilityInside
}

func interpolateUV(a, b UV, factor float64) UV {
	return UV{
		U: a.U + (b.U-a.U)*factor,
		V: a.V + (b.V-a.V)*factor,
	}
}

func (f *Frustum) ClipTriangle(
	pointsIn *[3]Vec4, uvsIn *[3]UV,
	pointsOut *[maxClipPoints][3]Vec4,
	uvsOut *[maxClipPoints][3]UV,
) (numOut int) {
	polygon := f.polygonPool.Get().(*Polygon)
	defer f.polygonPool.Put(polygon)
	polygon.Count = 0

	polygon2 := f.polygonPool.Get().(*Polygon)
	defer f.polygonPool.Put(polygon2)
	polygon2.Count = 0

	polygon.AddVertex(pointsIn[0], uvsIn[0])
	polygon.AddVertex(pointsIn[1], uvsIn[1])
	polygon.AddVertex(pointsIn[2], uvsIn[2])

	planes := []int{
		PlaneLeft,
		PlaneRight,
		PlaneTop,
		PlaneBottom,
		PlaneNear,
		PlaneFar,
	}

	// Iterate over each plane of the frustum and build a polygon from the input
	// triangle, containing only the vertices that are inside the frustum.
	for _, pi := range planes {
		plane := &f.Planes[pi]
		polygon2.Count = 0

		// Iterate over each edge of the polygon
		for b := 0; b < polygon.Count; b++ {
			a := (b + 1) % polygon.Count
			uvA, uvB := polygon.UVs[a], polygon.UVs[b]
			vertA, vertB := polygon.Points[a], polygon.Points[b]

			if plane.IsVertexInside(vertA) {
				if !plane.IsVertexInside(vertB) {
					intersect, factor := plane.Intersect(vertA, vertB)
					uv := interpolateUV(uvA, uvB, factor)
					polygon2.AddVertex(intersect, uv)
				}
				polygon2.AddVertex(vertA, uvA)
			} else if plane.IsVertexInside(vertB) {
				intersect, factor := plane.Intersect(vertA, vertB)
				uv := interpolateUV(uvA, uvB, factor)
				polygon2.AddVertex(intersect, uv)
			}
		}

		// All vertices are outside
		if polygon2.Count == 0 {
			return 0
		}

		polygon, polygon2 = polygon2, polygon
	}

	// Convert the polygon back to triangles
	return polygon.Triangulate(pointsOut, uvsOut)
}
