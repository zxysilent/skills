// Package geometry provides deterministic geometry primitives shared by the
// renderer and the SVG checker. Ported from draw_geometry.py.
package geometry

import "math"

// Point is a 2D coordinate.
type Point struct {
	X float64
	Y float64
}

// Bounds is an axis-aligned rectangle (left, top, right, bottom).
type Bounds struct {
	Left   float64
	Top    float64
	Right  float64
	Bottom float64
}

// EPSILON is the default float comparison tolerance.
const EPSILON = 1e-6

// EstimateTextWidth approximates the pixel width of rendered text at 12px font size.
func EstimateTextWidth(text string) float64 {
	// Rough estimate: average char ~7px, min 36px with 14px padding
	w := 0.0
	for _, r := range text {
		switch {
		case r >= 'A' && r <= 'Z':
			w += 8.5
		case r >= 'a' && r <= 'z':
			w += 7.0
		case r >= '0' && r <= '9':
			w += 7.5
		case r == ' ':
			w += 4.0
		case r > 127: // CJK/wide chars
			w += 14.0
		default:
			w += 5.5
		}
	}
	return w
}

// SegmentInteraction describes how two segments relate.
type SegmentInteraction struct {
	Kind          string // "crossing" | "touch" | "overlap"
	Point         Point
	HasPoint      bool
	OverlapLength float64
}

// RouteInteractions aggregates crossing/overlap facts for a route.
type RouteInteractions struct {
	Crossings    []Point
	OverlapCount int
	OverlapLen   float64
}

func AlmostEqual(first, second, epsilon float64) bool {
	return math.Abs(first-second) <= epsilon
}

func SamePoint(first, second Point, epsilon float64) bool {
	return AlmostEqual(first.X, second.X, epsilon) && AlmostEqual(first.Y, second.Y, epsilon)
}

// SegmentAxis returns "horizontal", "vertical", or "other".
func SegmentAxis(first, second Point) string {
	if AlmostEqual(first.Y, second.Y, EPSILON) {
		return "horizontal"
	}
	if AlmostEqual(first.X, second.X, EPSILON) {
		return "vertical"
	}
	return "other"
}

func RouteIsOrthogonal(points []Point) bool {
	for i := 0; i+1 < len(points); i++ {
		if SegmentAxis(points[i], points[i+1]) == "other" {
			return false
		}
	}
	return true
}

func RouteLength(points []Point) float64 {
	total := 0.0
	for i := 0; i+1 < len(points); i++ {
		total += math.Hypot(points[i+1].X-points[i].X, points[i+1].Y-points[i].Y)
	}
	return total
}

func BendCount(points []Point) int {
	axes := make([]string, 0, len(points))
	for i := 0; i+1 < len(points); i++ {
		if SamePoint(points[i], points[i+1], EPSILON) {
			continue
		}
		axes = append(axes, SegmentAxis(points[i], points[i+1]))
	}
	bends := 0
	for i := 0; i+1 < len(axes); i++ {
		if axes[i] != axes[i+1] {
			bends++
		}
	}
	return bends
}

func BoundsIntersect(first, second Bounds, padding float64) bool {
	return !(first.Right+padding <= second.Left ||
		second.Right+padding <= first.Left ||
		first.Bottom+padding <= second.Top ||
		second.Bottom+padding <= first.Top)
}

func ExpandBounds(bounds Bounds, padding float64) Bounds {
	return Bounds{bounds.Left - padding, bounds.Top - padding, bounds.Right + padding, bounds.Bottom + padding}
}

func PointInBounds(point Point, bounds Bounds, padding float64, interior bool) bool {
	if interior {
		return bounds.Left+padding < point.X && point.X < bounds.Right-padding &&
			bounds.Top+padding < point.Y && point.Y < bounds.Bottom-padding
	}
	return bounds.Left-padding <= point.X && point.X <= bounds.Right+padding &&
		bounds.Top-padding <= point.Y && point.Y <= bounds.Bottom+padding
}

func BoundsInside(inner, outer Bounds, padding float64) bool {
	return inner.Left >= outer.Left+padding &&
		inner.Top >= outer.Top+padding &&
		inner.Right <= outer.Right-padding &&
		inner.Bottom <= outer.Bottom-padding
}

func RouteInsideCanvas(points []Point, canvas Bounds, margin float64) bool {
	safe := Bounds{canvas.Left + margin, canvas.Top + margin, canvas.Right - margin, canvas.Bottom - margin}
	for _, p := range points {
		if !PointInBounds(p, safe, 0, false) {
			return false
		}
	}
	return true
}

func orientation(first, second, third Point) float64 {
	return (second.X-first.X)*(third.Y-first.Y) - (second.Y-first.Y)*(third.X-first.X)
}

func onSegment(point, first, second Point, epsilon float64) bool {
	return math.Min(first.X, second.X)-epsilon <= point.X && point.X <= math.Max(first.X, second.X)+epsilon &&
		math.Min(first.Y, second.Y)-epsilon <= point.Y && point.Y <= math.Max(first.Y, second.Y)+epsilon &&
		math.Abs(orientation(first, second, point)) <= epsilon
}

func uniquePoints(points []Point) []Point {
	unique := make([]Point, 0, len(points))
	for _, p := range points {
		found := false
		for _, u := range unique {
			if SamePoint(p, u, EPSILON) {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, p)
		}
	}
	return unique
}

// SegmentInteractionOf returns a proper crossing, touch, or collinear overlap
// for two segments, or nil when they do not interact.
func SegmentInteractionOf(a1, a2, b1, b2 Point) *SegmentInteraction {
	axisA := SegmentAxis(a1, a2)
	axisB := SegmentAxis(b1, b2)

	if axisA == "horizontal" && axisB == "horizontal" && AlmostEqual(a1.Y, b1.Y, EPSILON) {
		start := math.Max(math.Min(a1.X, a2.X), math.Min(b1.X, b2.X))
		end := math.Min(math.Max(a1.X, a2.X), math.Max(b1.X, b2.X))
		if end-start > EPSILON {
			return &SegmentInteraction{Kind: "overlap", OverlapLength: end - start}
		}
		if AlmostEqual(start, end, EPSILON) {
			return &SegmentInteraction{Kind: "touch", Point: Point{start, a1.Y}, HasPoint: true}
		}
		return nil
	}

	if axisA == "vertical" && axisB == "vertical" && AlmostEqual(a1.X, b1.X, EPSILON) {
		start := math.Max(math.Min(a1.Y, a2.Y), math.Min(b1.Y, b2.Y))
		end := math.Min(math.Max(a1.Y, a2.Y), math.Max(b1.Y, b2.Y))
		if end-start > EPSILON {
			return &SegmentInteraction{Kind: "overlap", OverlapLength: end - start}
		}
		if AlmostEqual(start, end, EPSILON) {
			return &SegmentInteraction{Kind: "touch", Point: Point{a1.X, start}, HasPoint: true}
		}
		return nil
	}

	if axisA == "horizontal" && axisB == "vertical" {
		point := Point{b1.X, a1.Y}
		if onSegment(point, a1, a2, EPSILON) && onSegment(point, b1, b2, EPSILON) {
			return &SegmentInteraction{Kind: "crossing", Point: point, HasPoint: true}
		}
		return nil
	}

	if axisA == "vertical" && axisB == "horizontal" {
		point := Point{a1.X, b1.Y}
		if onSegment(point, a1, a2, EPSILON) && onSegment(point, b1, b2, EPSILON) {
			return &SegmentInteraction{Kind: "crossing", Point: point, HasPoint: true}
		}
		return nil
	}

	// General line intersection for authored SVGs and sampled curves.
	o1 := orientation(a1, a2, b1)
	o2 := orientation(a1, a2, b2)
	o3 := orientation(b1, b2, a1)
	o4 := orientation(b1, b2, a2)
	if math.Abs(o1) <= EPSILON && math.Abs(o2) <= EPSILON && math.Abs(o3) <= EPSILON && math.Abs(o4) <= EPSILON {
		candidates := make([]Point, 0, 4)
		for _, p := range []Point{a1, a2, b1, b2} {
			if onSegment(p, a1, a2, EPSILON) && onSegment(p, b1, b2, EPSILON) {
				candidates = append(candidates, p)
			}
		}
		unique := uniquePoints(candidates)
		if len(unique) >= 2 {
			maxDist := 0.0
			for _, a := range unique {
				for _, b := range unique {
					d := math.Hypot(a.X-b.X, a.Y-b.Y)
					if d > maxDist {
						maxDist = d
					}
				}
			}
			return &SegmentInteraction{Kind: "overlap", OverlapLength: maxDist}
		}
		if len(unique) == 1 {
			return &SegmentInteraction{Kind: "touch", Point: unique[0], HasPoint: true}
		}
		return nil
	}
	if o1*o2 <= EPSILON && o3*o4 <= EPSILON {
		denominator := (a1.X-a2.X)*(b1.Y-b2.Y) - (a1.Y-a2.Y)*(b1.X-b2.X)
		if math.Abs(denominator) <= EPSILON {
			return nil
		}
		detFirst := a1.X*a2.Y - a1.Y*a2.X
		detSecond := b1.X*b2.Y - b1.Y*b2.X
		x := (detFirst*(b1.X-b2.X) - (a1.X-a2.X)*detSecond) / denominator
		y := (detFirst*(b1.Y-b2.Y) - (a1.Y-a2.Y)*detSecond) / denominator
		point := Point{x, y}
		if onSegment(point, a1, a2, EPSILON) && onSegment(point, b1, b2, EPSILON) {
			return &SegmentInteraction{Kind: "crossing", Point: point, HasPoint: true}
		}
	}
	return nil
}

// RouteInteractionsOf computes how a route interacts with existing routes.
func RouteInteractionsOf(route []Point, others [][]Point) RouteInteractions {
	result := RouteInteractions{Crossings: []Point{}}
	for _, other := range others {
		for i := 0; i+1 < len(route); i++ {
			for j := 0; j+1 < len(other); j++ {
				interaction := SegmentInteractionOf(route[i], route[i+1], other[j], other[j+1])
				if interaction == nil {
					continue
				}
				switch interaction.Kind {
				case "crossing":
					result.Crossings = append(result.Crossings, interaction.Point)
				case "overlap":
					result.OverlapCount++
					result.OverlapLen += interaction.OverlapLength
				}
			}
		}
	}
	return result
}
