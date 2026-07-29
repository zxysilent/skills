// Package route provides deterministic orthogonal route planning.
// Ported from generate-from-template.py: build_orthogonal_route + visibility_grid_route.
package route

import (
	"container/heap"
	"fmt"
	"math"
	"sort"

	"drawcli/geometry"
)

// Point is a 2D coordinate.
type Point = geometry.Point

// Bounds is an axis-aligned rectangle.
type Bounds = geometry.Bounds

type routeScorer struct {
	HintX          []float64
	HintY          []float64
	SourcePort     string
	TargetPort     string
	ExistingRoutes [][]Point
	PortAxis       func(string) string // "horizontal" or "vertical"
}

func newScorer(hintX, hintY []float64, srcPort, tgtPort string, existing [][]Point) *routeScorer {
	portAxis := func(port string) string {
		switch port {
		case "left", "right":
			return "horizontal"
		case "top", "bottom":
			return "vertical"
		}
		return ""
	}
	return &routeScorer{
		HintX: hintX, HintY: hintY,
		SourcePort: srcPort, TargetPort: tgtPort,
		ExistingRoutes: existing,
		PortAxis:       portAxis,
	}
}

func routeLength(pts []Point) float64 {
	total := 0.0
	for i := 0; i+1 < len(pts); i++ {
		total += math.Abs(pts[i+1].X-pts[i].X) + math.Abs(pts[i+1].Y-pts[i].Y)
	}
	return total
}

func bendCount(pts []Point) int {
	if len(pts) < 3 {
		return 0
	}
	bends := 0
	for i := 1; i+1 < len(pts); i++ {
		dx1 := pts[i].X - pts[i-1].X
		dy1 := pts[i].Y - pts[i-1].Y
		dx2 := pts[i+1].X - pts[i].X
		dy2 := pts[i+1].Y - pts[i].Y
		axis1 := "horizontal"
		if math.Abs(dy1) > math.Abs(dx1) {
			axis1 = "vertical"
		}
		axis2 := "horizontal"
		if math.Abs(dy2) > math.Abs(dx2) {
			axis2 = "vertical"
		}
		if axis1 != axis2 {
			bends++
		}
	}
	return bends
}

func usesLane(pts []Point, value float64, axis string, tolerance float64) bool {
	for _, p := range pts {
		if axis == "x" && math.Abs(p.X-value) <= tolerance {
			return true
		}
		if axis != "x" && math.Abs(p.Y-value) <= tolerance {
			return true
		}
	}
	return false
}

func (s *routeScorer) score(pts []Point) float64 {
	length := routeLength(pts)
	bends := bendCount(pts)
	score := length + float64(bends)*22

	if len(pts) >= 2 && s.SourcePort != "" {
		firstAxis := segmentAxis(pts[0], pts[1])
		if firstAxis != s.PortAxis(s.SourcePort) {
			score += 180
		}
	}
	if len(pts) >= 2 && s.TargetPort != "" {
		lastAxis := segmentAxis(pts[len(pts)-2], pts[len(pts)-1])
		if lastAxis != s.PortAxis(s.TargetPort) {
			score += 180
		}
	}
	for _, lane := range s.HintX {
		if usesLane(pts, lane, "x", 1) {
			score -= 28
		}
	}
	for _, lane := range s.HintY {
		if usesLane(pts, lane, "y", 1) {
			score -= 28
		}
	}
	interactions := geometry.RouteInteractionsOf(pts, s.ExistingRoutes)
	score += float64(len(interactions.Crossings)) * 640
	score += float64(interactions.OverlapCount)*900 + interactions.OverlapLen*18
	return score
}

func segmentAxis(a, b Point) string {
	if math.Abs(a.Y-b.Y) < 1e-6 {
		return "horizontal"
	}
	if math.Abs(a.X-b.X) < 1e-6 {
		return "vertical"
	}
	return "other"
}

func routeIsOrthogonal(pts []Point) bool {
	for i := 0; i+1 < len(pts); i++ {
		if segmentAxis(pts[i], pts[i+1]) == "other" {
			return false
		}
	}
	return true
}

func simplifyPoints(pts []Point, protected []Point) []Point {
	protectedSet := make(map[string]bool)
	for _, p := range protected {
		protectedSet[fmt.Sprintf("%.2f,%.2f", p.X, p.Y)] = true
	}

	rounded := make([]Point, 0, len(pts))
	for _, p := range pts {
		rp := Point{X: math.Round(p.X*100) / 100, Y: math.Round(p.Y*100) / 100}
		if len(rounded) > 0 && rounded[len(rounded)-1] == rp {
			continue
		}
		rounded = append(rounded, rp)
	}

	collapsed := make([]Point, 0, len(rounded))
	for _, p := range rounded {
		if len(collapsed) < 2 {
			collapsed = append(collapsed, p)
			continue
		}
		p0, p1 := collapsed[len(collapsed)-2], collapsed[len(collapsed)-1]
		key := fmt.Sprintf("%.2f,%.2f", p1.X, p1.Y)
		colVert := math.Abs(p0.X-p1.X) < 1e-6 && math.Abs(p1.X-p.X) < 1e-6
		colHoriz := math.Abs(p0.Y-p1.Y) < 1e-6 && math.Abs(p1.Y-p.Y) < 1e-6
		if (colVert || colHoriz) && !protectedSet[key] {
			collapsed[len(collapsed)-1] = p
		} else {
			collapsed = append(collapsed, p)
		}
	}
	return collapsed
}

// segmentHitsBounds returns true if the line segment p1-p2 crosses into bounds.
func segmentHitsBounds(p1, p2 Point, b Bounds) bool {
	eps := 1e-6
	if math.Abs(p1.Y-p2.Y) < eps {
		y := p1.Y
		if !(b.Top+eps < y && y < b.Bottom-eps) {
			return false
		}
		segL, segR := math.Min(p1.X, p2.X), math.Max(p1.X, p2.X)
		ol := math.Max(segL, b.Left)
		or := math.Min(segR, b.Right)
		if or-ol <= eps {
			return false
		}
		if (math.Abs(ol-p1.X) < eps && math.Abs(or-p1.X) < eps) ||
			(math.Abs(ol-p2.X) < eps && math.Abs(or-p2.X) < eps) {
			return false
		}
		return true
	}
	if math.Abs(p1.X-p2.X) < eps {
		x := p1.X
		if !(b.Left+eps < x && x < b.Right-eps) {
			return false
		}
		segT, segB := math.Min(p1.Y, p2.Y), math.Max(p1.Y, p2.Y)
		ot := math.Max(segT, b.Top)
		ob := math.Min(segB, b.Bottom)
		if ob-ot <= eps {
			return false
		}
		if (math.Abs(ot-p1.Y) < eps && math.Abs(ob-p1.Y) < eps) ||
			(math.Abs(ot-p2.Y) < eps && math.Abs(ob-p2.Y) < eps) {
			return false
		}
		return true
	}
	return false
}

func collisionCount(pts []Point, obstacles []Bounds) int {
	count := 0
	for i := 0; i+1 < len(pts); i++ {
		for _, obs := range obstacles {
			if segmentHitsBounds(pts[i], pts[i+1], obs) {
				count++
			}
		}
	}
	return count
}

func routeCollides(pts []Point, obstacles []Bounds) bool {
	return collisionCount(pts, obstacles) > 0
}

func expandBounds(b Bounds, pad float64) Bounds {
	return Bounds{b.Left - pad, b.Top - pad, b.Right + pad, b.Bottom + pad}
}

func pointInBounds(p Point, b Bounds, interior bool) bool {
	if interior {
		return b.Left < p.X && p.X < b.Right && b.Top < p.Y && p.Y < b.Bottom
	}
	return b.Left-1e-6 <= p.X && p.X <= b.Right+1e-6 && b.Top-1e-6 <= p.Y && p.Y <= b.Bottom+1e-6
}

func routeInsideCanvas(pts []Point, canvas Bounds, margin float64) bool {
	safe := Bounds{canvas.Left + margin, canvas.Top + margin, canvas.Right - margin, canvas.Bottom - margin}
	for _, p := range pts {
		if !pointInBounds(p, safe, false) {
			return false
		}
	}
	return true
}

func offsetPoint(p Point, port string, dist float64) Point {
	switch port {
	case "left":
		return Point{p.X - dist, p.Y}
	case "right":
		return Point{p.X + dist, p.Y}
	case "top":
		return Point{p.X, p.Y - dist}
	case "bottom":
		return Point{p.X, p.Y + dist}
	}
	return p
}

func clearPortPoint(endpoint Point, port string, desired float64, obstacles []Bounds, canvas *Bounds) Point {
	fractions := []float64{1.0, 0.75, 0.5, 0.35, 0.2, 0.0}
	for _, frac := range fractions {
		dist := desired * frac
		candidate := offsetPoint(endpoint, port, dist)
		if canvas != nil && !pointInBounds(candidate, *canvas, false) {
			continue
		}
		blocked := false
		for _, obs := range obstacles {
			if pointInBounds(candidate, obs, true) || segmentHitsBounds(endpoint, candidate, obs) {
				blocked = true
				break
			}
		}
		if !blocked {
			return candidate
		}
	}
	return endpoint
}

// priorityQueue for A*
type pqItem struct {
	cost  float64
	path  []Point
	point Point
	axis  string
	index int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].cost < pq[j].cost || (math.Abs(pq[i].cost-pq[j].cost) < 1e-6 && len(pq[i].path) < len(pq[j].path))
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	item := x.(*pqItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// visibilityGridRoute finds a rectilinear route on an obstacle visibility grid (A*).
func visibilityGridRoute(start, end Point, obstacles []Bounds, canvas *Bounds, hintX, hintY []float64, existingRoutes [][]Point) []Point {
	if start == end {
		return []Point{start}
	}

	var canvasBounds Bounds
	if canvas == nil {
		allX := []float64{start.X, end.X}
		allY := []float64{start.Y, end.Y}
		for _, b := range obstacles {
			allX = append(allX, b.Left, b.Right)
			allY = append(allY, b.Top, b.Bottom)
		}
		sort.Float64s(allX)
		sort.Float64s(allY)
		canvasBounds = Bounds{allX[0] - 64, allY[0] - 64, allX[len(allX)-1] + 64, allY[len(allY)-1] + 64}
	} else {
		canvasBounds = *canvas
	}

	inset := 4.0
	xSet := map[float64]bool{
		start.X: true, end.X: true,
		(start.X + end.X) / 2:      true,
		canvasBounds.Left + inset:  true,
		canvasBounds.Right - inset: true,
	}
	for _, v := range hintX {
		xSet[math.Round(v*100)/100] = true
	}
	ySet := map[float64]bool{
		start.Y: true, end.Y: true,
		(start.Y + end.Y) / 2:       true,
		canvasBounds.Top + inset:    true,
		canvasBounds.Bottom - inset: true,
	}
	for _, v := range hintY {
		ySet[math.Round(v*100)/100] = true
	}
	for _, b := range obstacles {
		xSet[math.Round(b.Left*100)/100] = true
		xSet[math.Round(b.Right*100)/100] = true
		ySet[math.Round(b.Top*100)/100] = true
		ySet[math.Round(b.Bottom*100)/100] = true
	}
	for _, route := range existingRoutes {
		for _, p := range route {
			for _, d := range []float64{-10, 0, 10} {
				xSet[math.Round((p.X+d)*100)/100] = true
				ySet[math.Round((p.Y+d)*100)/100] = true
			}
		}
	}

	// Filter to canvas bounds
	var xs, ys []float64
	for x := range xSet {
		if canvasBounds.Left-1e-6 <= x && x <= canvasBounds.Right+1e-6 {
			xs = append(xs, x)
		}
	}
	for y := range ySet {
		if canvasBounds.Top-1e-6 <= y && y <= canvasBounds.Bottom+1e-6 {
			ys = append(ys, y)
		}
	}
	sort.Float64s(xs)
	sort.Float64s(ys)

	// Build visibility points (outside all obstacles)
	points := make(map[Point]bool)
	for _, x := range xs {
		for _, y := range ys {
			p := Point{x, y}
			inside := false
			for _, obs := range obstacles {
				if pointInBounds(p, obs, true) {
					inside = true
					break
				}
			}
			if !inside {
				points[p] = true
			}
		}
	}
	points[start] = true
	points[end] = true

	// Group points by Y and X for adjacency
	byY := make(map[float64][]Point)
	byX := make(map[float64][]Point)
	for p := range points {
		byY[p.Y] = append(byY[p.Y], p)
		byX[p.X] = append(byX[p.X], p)
	}

	adj := make(map[Point][]Point)
	for p := range points {
		adj[p] = nil
	}

	connect := func(line []Point, sortKey int) {
		sort.Slice(line, func(i, j int) bool {
			if sortKey == 0 {
				return line[i].X < line[j].X || (math.Abs(line[i].X-line[j].X) < 1e-6 && line[i].Y < line[j].Y)
			}
			return line[i].Y < line[j].Y || (math.Abs(line[i].Y-line[j].Y) < 1e-6 && line[i].X < line[j].X)
		})
		for i := 0; i+1 < len(line); i++ {
			blocked := false
			for _, obs := range obstacles {
				if segmentHitsBounds(line[i], line[i+1], obs) {
					blocked = true
					break
				}
			}
			if !blocked {
				adj[line[i]] = append(adj[line[i]], line[i+1])
				adj[line[i+1]] = append(adj[line[i+1]], line[i])
			}
		}
	}

	for _, line := range byY {
		connect(line, 0)
	}
	for _, line := range byX {
		connect(line, 1)
	}

	// A* with bend cost
	type stateKey struct {
		point Point
		axis  string
	}
	startState := stateKey{start, ""}
	distances := map[stateKey]float64{startState: 0}
	paths := map[stateKey][]Point{startState: {start}}

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{cost: 0, path: []Point{start}, point: start, axis: ""})

	var best *pqItem
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*pqItem)
		sk := stateKey{item.point, item.axis}
		if c, ok := distances[sk]; ok && item.cost > c+1e-6 {
			continue
		}

		if item.point == end {
			best = item
			break
		}

		for _, nb := range adj[item.point] {
			axis := segmentAxis(item.point, nb)
			if axis == "other" {
				continue
			}
			dist := math.Abs(nb.X-item.point.X) + math.Abs(nb.Y-item.point.Y)
			extra := dist
			if item.axis != "" && item.axis != axis {
				extra += 22
			}
			segRoute := []Point{item.point, nb}
			interactions := geometry.RouteInteractionsOf(segRoute, existingRoutes)
			extra += float64(len(interactions.Crossings)) * 640
			extra += float64(interactions.OverlapCount)*10000 + interactions.OverlapLen*30

			if axis == "vertical" {
				for _, v := range hintX {
					if math.Abs(item.point.X-v) <= 1 {
						extra -= math.Min(18, dist*0.08)
					}
				}
			}
			if axis == "horizontal" {
				for _, v := range hintY {
					if math.Abs(item.point.Y-v) <= 1 {
						extra -= math.Min(18, dist*0.08)
					}
				}
			}

			nextState := stateKey{nb, axis}
			nextCost := item.cost + extra
			nextPath := append([]Point{}, item.path...)
			nextPath = append(nextPath, nb)

			oldCost, hasOld := distances[nextState]
			if !hasOld || nextCost < oldCost-1e-6 {
				distances[nextState] = nextCost
				paths[nextState] = nextPath
				heap.Push(pq, &pqItem{cost: nextCost, path: nextPath, point: nb, axis: axis})
			}
		}
	}

	if best == nil {
		return nil
	}
	return simplifyPoints(best.path, nil)
}

// OrthogonalRoute computes a collision-free orthogonal route from start to end.
func OrthogonalRoute(
	start, end Point,
	obstacles []Bounds,
	sourcePort, targetPort string,
	routePoints [][2]float64,
	hintX, hintY []float64,
	routingPadding, portClearance float64,
	canvasBounds *Bounds,
	existingRoutes [][]Point,
) ([]Point, error) {

	// If explicit waypoints are given, route through them
	if len(routePoints) >= 2 {
		waypoints := make([]Point, len(routePoints))
		for i, rp := range routePoints {
			waypoints[i] = Point{rp[0], rp[1]}
		}
		mandatory := append([]Point{start}, waypoints...)
		mandatory = append(mandatory, end)
		direct := simplifyPoints(mandatory, waypoints)
		if routeIsOrthogonal(direct) && !routeCollides(direct, obstacles) && (canvasBounds == nil || routeInsideCanvas(direct, *canvasBounds, 0)) {
			return direct, nil
		}

		assembled := make([]Point, 0)
		for i := 0; i+1 < len(mandatory); i++ {
			segSrc := mandatory[i]
			segDst := mandatory[i+1]
			srcPort, tgtPort := sourcePort, targetPort
			if i > 0 {
				srcPort = ""
			}
			if i < len(mandatory)-2 {
				tgtPort = ""
			}
			occ := make([][]Point, len(existingRoutes))
			copy(occ, existingRoutes)
			if len(assembled) >= 2 {
				occ = append(occ, assembled)
			}
			seg, err := OrthogonalRoute(segSrc, segDst, obstacles, srcPort, tgtPort, nil, hintX, hintY, routingPadding, portClearance, canvasBounds, occ)
			if err != nil {
				return nil, fmt.Errorf("segment %d: %w", i, err)
			}
			if len(assembled) > 0 {
				assembled = append(assembled, seg[1:]...)
			} else {
				assembled = append(assembled, seg...)
			}
		}
		result := simplifyPoints(assembled, waypoints)
		if !routeIsOrthogonal(result) {
			return nil, fmt.Errorf("explicit route waypoints could not be connected orthogonally")
		}
		if routeCollides(result, obstacles) {
			return nil, fmt.Errorf("explicit route waypoints cannot be connected without crossing an obstacle")
		}
		return result, nil
	}

	if routingPadding <= 0 {
		routingPadding = 24
	}
	if portClearance <= 0 {
		portClearance = math.Max(18, routingPadding*0.85)
	}

	srcPortNorm := sourcePort
	tgtPortNorm := targetPort

	innerStart := clearPortPoint(start, srcPortNorm, portClearance, obstacles, canvasBounds)
	innerEnd := clearPortPoint(end, tgtPortNorm, portClearance, obstacles, canvasBounds)

	ssx, ssy := innerStart.X, innerStart.Y
	eex, eey := innerEnd.X, innerEnd.Y

	// Expanded obstacles with clearance
	expanded := make([]Bounds, len(obstacles))
	for i, b := range obstacles {
		padded := expandBounds(b, routingPadding)
		// If a clearance point is inside the padded halo but not inside the real obstacle, use the real one
		insideHalo := false
		for _, p := range []Point{start, end, innerStart, innerEnd} {
			if pointInBounds(p, padded, true) && !pointInBounds(p, b, true) {
				insideHalo = true
				break
			}
		}
		if insideHalo {
			expanded[i] = b
		} else {
			expanded[i] = padded
		}
	}

	// Collect lane candidates
	laneXSet := map[float64]bool{ssx: true, eex: true, (ssx + eex) / 2: true}
	laneYSet := map[float64]bool{ssy: true, eey: true, (ssy + eey) / 2: true}
	for _, v := range hintX {
		laneXSet[v] = true
	}
	for _, v := range hintY {
		laneYSet[v] = true
	}
	for _, b := range expanded {
		laneXSet[b.Left] = true
		laneXSet[b.Right] = true
		laneYSet[b.Top] = true
		laneYSet[b.Bottom] = true
	}

	var laneX, laneY []float64
	for v := range laneXSet {
		laneX = append(laneX, v)
	}
	for v := range laneYSet {
		laneY = append(laneY, v)
	}
	sort.Float64s(laneX)
	sort.Float64s(laneY)

	// Calculate rails (outer boundaries)
	var leftRail, rightRail, topRail, bottomRail float64
	if len(expanded) > 0 {
		leftRail, rightRail = expanded[0].Left, expanded[0].Right
		topRail, bottomRail = expanded[0].Top, expanded[0].Bottom
		for _, b := range expanded {
			leftRail = math.Min(leftRail, b.Left)
			rightRail = math.Max(rightRail, b.Right)
			topRail = math.Min(topRail, b.Top)
			bottomRail = math.Max(bottomRail, b.Bottom)
		}
		leftRail -= 24
		rightRail += 24
		topRail -= 24
		bottomRail += 24
	} else {
		leftRail = math.Min(ssx, eex) - 48
		rightRail = math.Max(ssx, eex) + 48
		topRail = math.Min(ssy, eey) - 48
		bottomRail = math.Max(ssy, eey) + 48
	}

	// Build candidate routes
	type candidate struct {
		pts []Point
	}

	var candidates []candidate
	addCand := func(pts []Point) {
		candidates = append(candidates, candidate{pts})
	}

	addCand([]Point{start, innerStart, innerEnd, end})
	addCand([]Point{start, innerStart, {eex, ssy}, innerEnd, end})
	addCand([]Point{start, innerStart, {ssx, eey}, innerEnd, end})
	addCand([]Point{start, innerStart, {(ssx + eex) / 2, ssy}, {(ssx + eex) / 2, eey}, innerEnd, end})
	addCand([]Point{start, innerStart, {ssx, (ssy + eey) / 2}, {eex, (ssy + eey) / 2}, innerEnd, end})
	addCand([]Point{start, innerStart, {leftRail, ssy}, {leftRail, eey}, innerEnd, end})
	addCand([]Point{start, innerStart, {rightRail, ssy}, {rightRail, eey}, innerEnd, end})
	addCand([]Point{start, innerStart, {ssx, topRail}, {eex, topRail}, innerEnd, end})
	addCand([]Point{start, innerStart, {ssx, bottomRail}, {eex, bottomRail}, innerEnd, end})

	for _, x := range laneX {
		addCand([]Point{start, innerStart, {x, ssy}, {x, eey}, innerEnd, end})
	}
	for _, y := range laneY {
		addCand([]Point{start, innerStart, {ssx, y}, {eex, y}, innerEnd, end})
	}
	for _, x := range hintX {
		for _, y := range hintY {
			addCand([]Point{start, innerStart, {x, ssy}, {x, y}, {eex, y}, innerEnd, end})
		}
	}

	// Also run visibility grid
	vis := visibilityGridRoute(innerStart, innerEnd, expanded, canvasBounds, hintX, hintY, existingRoutes)
	if vis != nil {
		vpts := append([]Point{start}, vis...)
		vpts = append(vpts, end)
		addCand(vpts)
	}

	// Score and pick best
	scorer := newScorer(hintX, hintY, srcPortNorm, tgtPortNorm, existingRoutes)

	var bestRoute []Point
	bestScore := math.MaxFloat64

	for _, cand := range candidates {
		simplified := simplifyPoints(cand.pts, nil)
		if !routeIsOrthogonal(simplified) {
			continue
		}
		if canvasBounds != nil && !routeInsideCanvas(simplified, *canvasBounds, 0) {
			continue
		}
		coll := collisionCount(simplified, expanded)
		scr := scorer.score(simplified)
		if coll == 0 && scr < bestScore {
			bestScore = scr
			bestRoute = simplified
		}
	}

	if bestRoute != nil {
		return bestRoute, nil
	}
	return nil, fmt.Errorf("no collision-free orthogonal route found")
}
