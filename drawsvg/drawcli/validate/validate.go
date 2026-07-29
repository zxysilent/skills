// Package validate provides SVG syntax, marker, geometry, and composition checks.
// Ported from validate_svg.py (961 lines) + composition_quality.py (271 lines).
package validate

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"drawcli/geometry"
)

// --- XML and Marker Checks ---

type CheckResult struct {
	Ok      bool
	Details map[string]any
}

// CheckXML validates SVG XML syntax and tag balance.
func CheckXML(svgContent string) CheckResult {
	details := make(map[string]any)
	decoder := xml.NewDecoder(strings.NewReader(svgContent))
	tagStack := make([]string, 0)
	elementCount := 0

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			tagStack = append(tagStack, t.Name.Local)
			elementCount++
		case xml.EndElement:
			if len(tagStack) == 0 {
				details["error"] = fmt.Sprintf("unexpected closing tag </%s>", t.Name.Local)
				return CheckResult{Ok: false, Details: details}
			}
			last := tagStack[len(tagStack)-1]
			if last != t.Name.Local {
				details["error"] = fmt.Sprintf("tag mismatch: <%s> closed by </%s>", last, t.Name.Local)
				return CheckResult{Ok: false, Details: details}
			}
			tagStack = tagStack[:len(tagStack)-1]
		}
	}

	if len(tagStack) > 0 {
		details["error"] = fmt.Sprintf("unclosed tags: %v", tagStack)
		return CheckResult{Ok: false, Details: details}
	}

	details["elements"] = elementCount
	return CheckResult{Ok: true, Details: details}
}

// CheckMarkers validates that all url(#...) references resolve to definitions.
func CheckMarkers(svgContent string) CheckResult {
	details := make(map[string]any)

	// Find all url(#id) references
	refRegex := regexp.MustCompile(`url\(#([^)]+)\)`)
	refMatches := refRegex.FindAllStringSubmatch(svgContent, -1)

	// Find all id="..." definitions
	defRegex := regexp.MustCompile(`id="([^"]+)"`)
	defMatches := defRegex.FindAllStringSubmatch(svgContent, -1)

	defs := make(map[string]bool)
	for _, m := range defMatches {
		defs[m[1]] = true
	}

	unresolved := make([]string, 0)
	seen := make(map[string]bool)
	for _, m := range refMatches {
		id := m[1]
		if seen[id] {
			continue
		}
		seen[id] = true
		if !defs[id] {
			unresolved = append(unresolved, id)
		}
	}

	details["references"] = len(seen)
	details["definitions"] = len(defs)

	if len(unresolved) > 0 {
		details["unresolved"] = unresolved
		return CheckResult{Ok: false, Details: details}
	}

	return CheckResult{Ok: true, Details: details}
}

// CheckGeometry validates arrow-component collision and route budgets.
func CheckGeometry(svgContent string) CheckResult {
	details := make(map[string]any)
	issues := make([]string, 0)

	// Check viewBox exists and is non-zero
	vbRegex := regexp.MustCompile(`viewBox="([^"]+)"`)
	vbMatch := vbRegex.FindStringSubmatch(svgContent)
	if vbMatch == nil {
		issues = append(issues, "missing viewBox")
	} else {
		details["viewBox"] = vbMatch[1]
	}

	// Check no path with implausible coordinates (common AI artifact)
	largeCoord := regexp.MustCompile(`L\s+(\d{5,})`)
	if largeCoord.MatchString(svgContent) {
		issues = append(issues, "implausibly large path coordinates")
	}

	// Check filtered elements are not too close to viewBox edge
	if len(vbMatch) >= 2 {
		var vx, vy, vw, vh float64
		fmt.Sscanf(vbMatch[1], "%f %f %f %f", &vx, &vy, &vw, &vh)
		filterElem := regexp.MustCompile(`<[^>]*filter="[^"]*"[^>]*x="([^"]*)"[^>]*y="([^"]*)"`)
		filterMatches := filterElem.FindAllStringSubmatch(svgContent, -1)
		for _, m := range filterMatches {
			var fx, fy float64
			fmt.Sscanf(m[1], "%f", &fx)
			fmt.Sscanf(m[2], "%f", &fy)
			if fx < vx+30 || fy < vy+30 || fx > vx+vw-60 || fy > vy+vh-60 {
				issues = append(issues, "filtered element near viewBox edge")
				break
			}
		}
	}

	if len(issues) > 0 {
		details["issues"] = issues
		return CheckResult{Ok: false, Details: details}
	}
	shapes, paths, err := svgGeometry(svgContent)
	if err != nil {
		details["issues"] = []string{err.Error()}
		return CheckResult{Ok: false, Details: details}
	}
	for _, path := range paths {
		for _, shape := range shapes {
			for index := 0; index+1 < len(path.route); index++ {
				if segmentHitsInterior(path.route[index], path.route[index+1], shape.bounds) {
					issues = append(issues, fmt.Sprintf("edge_node: path#%s intersects rect#%s", path.id, shape.id))
					break
				}
			}
		}
	}
	for first := 0; first < len(paths); first++ {
		for second := first + 1; second < len(paths); second++ {
			interaction := geometry.RouteInteractionsOf(paths[first].route, [][]geometry.Point{paths[second].route})
			if interaction.OverlapCount > 0 {
				issues = append(issues, fmt.Sprintf("edge_overlap: path#%s overlaps path#%s", paths[first].id, paths[second].id))
			}
		}
	}
	if len(issues) > 0 {
		details["issues"] = issues
		return CheckResult{Ok: false, Details: details}
	}
	return CheckResult{Ok: true, Details: details}
}

// CheckCollisions reports only edge-to-node intersections, matching the
// standalone Python collisions check without applying broader route budgets.
func CheckCollisions(svgContent string) CheckResult {
	shapes, paths, err := svgGeometry(svgContent)
	if err != nil {
		return CheckResult{Ok: false, Details: map[string]any{"issues": []string{err.Error()}}}
	}
	issues := make([]string, 0)
	for _, path := range paths {
		for _, shape := range shapes {
			for index := 0; index+1 < len(path.route); index++ {
				if segmentHitsInterior(path.route[index], path.route[index+1], shape.bounds) {
					issues = append(issues, fmt.Sprintf("edge_node: path#%s intersects rect#%s", path.id, shape.id))
					break
				}
			}
		}
	}
	return CheckResult{Ok: len(issues) == 0, Details: map[string]any{"issues": issues}}
}

// CheckComposition validates composition quality budgets.
func CheckComposition(svgContent string) CheckResult {
	contract, nodes, containers, labels, edges, err := compositionInput(svgContent)
	if err != nil {
		return CheckResult{Ok: false, Details: map[string]any{"issues": []string{err.Error()}}}
	}
	issues := make([]string, 0)
	minimumNodeGap := math.Inf(1)
	for first := 0; first < len(nodes); first++ {
		for second := first + 1; second < len(nodes); second++ {
			gap := compositionRectangleGap(nodes[first].bounds, nodes[second].bounds)
			minimumNodeGap = math.Min(minimumNodeGap, gap)
			if gap+1e-6 < contract.minNodeGap {
				issues = append(issues, fmt.Sprintf("composition_node_gap: %s,%s actual=%.2f limit=%.2f", nodes[first].id, nodes[second].id, gap, contract.minNodeGap))
			}
		}
	}
	minimumContainerGutter := math.Inf(1)
	for _, node := range nodes {
		container, ok := compositionContainingContainer(node.bounds, containers)
		if !ok {
			continue
		}
		gutter := compositionContainerGutter(node.bounds, container.bounds)
		minimumContainerGutter = math.Min(minimumContainerGutter, gutter)
		if gutter+1e-6 < contract.minContainerGutter {
			issues = append(issues, fmt.Sprintf("composition_container_gutter: %s@%s actual=%.2f limit=%.2f", node.id, container.id, gutter, contract.minContainerGutter))
		}
	}
	for _, label := range labels {
		expanded := compositionExpandBounds(label.bounds, contract.minLabelClearance)
		for _, node := range nodes {
			if compositionBoundsIntersect(expanded, node.bounds) {
				issues = append(issues, fmt.Sprintf("composition_label_clearance: %s is too close to node %s", label.id, node.id))
				break
			}
		}
		for _, edge := range edges {
			if edge.id == label.owner {
				continue
			}
			if compositionRouteHitsBounds(edge.route, expanded) {
				issues = append(issues, fmt.Sprintf("composition_label_clearance: %s is too close to edge %s", label.id, edge.id))
			}
		}
	}
	for first := 0; first < len(labels); first++ {
		expanded := compositionExpandBounds(labels[first].bounds, contract.minLabelClearance)
		for second := first + 1; second < len(labels); second++ {
			if compositionBoundsIntersect(expanded, labels[second].bounds) {
				issues = append(issues, fmt.Sprintf("composition_label_clearance: %s is too close to label %s", labels[first].id, labels[second].id))
			}
		}
	}
	totalBends := 0
	totalBridges := 0
	maxBends := 0
	for _, edge := range edges {
		totalBridges += edge.bridges
		bends := countBendsPoints(edge.route)
		totalBends += bends
		maxBends = max(maxBends, bends)
		if bends > contract.maxBendsPerEdge {
			issues = append(issues, fmt.Sprintf("composition_edge_bend_budget: %s actual=%d limit=%d", edge.id, bends, contract.maxBendsPerEdge))
		}
		stretch := compositionRouteStretch(edge.route)
		if stretch > contract.maxRouteStretch+1e-6 {
			issues = append(issues, fmt.Sprintf("composition_edge_route_stretch: %s actual=%.3f limit=%.3f", edge.id, stretch, contract.maxRouteStretch))
		}
		if shortest := compositionShortestSegment(edge.route); shortest >= 0 && shortest+1e-6 < contract.minSegmentLength {
			issues = append(issues, fmt.Sprintf("composition_edge_micro_segment: %s actual=%.2f limit=%.2f", edge.id, shortest, contract.minSegmentLength))
		}
	}
	if totalBends > contract.maxTotalBends {
		issues = append(issues, fmt.Sprintf("composition_total_bend_budget: diagram actual=%d limit=%d", totalBends, contract.maxTotalBends))
	}
	if totalBridges > contract.maxBridgedCrossings {
		issues = append(issues, fmt.Sprintf("composition_bridge_budget: diagram actual=%d limit=%d", totalBridges, contract.maxBridgedCrossings))
	}
	details := map[string]any{
		"max_bends_per_edge": maxBends,
		"total_bends":        totalBends,
		"bridged_crossings":  totalBridges,
		"profile":            contract.profile,
	}
	if !math.IsInf(minimumNodeGap, 1) {
		details["minimum_node_gap"] = minimumNodeGap
	}
	if !math.IsInf(minimumContainerGutter, 1) {
		details["minimum_container_gutter"] = minimumContainerGutter
	}
	if len(issues) > 0 {
		details["issues"] = issues
	}
	return CheckResult{Ok: len(issues) == 0, Details: details}
}

type svgCompositionContract struct {
	profile             string
	maxBendsPerEdge     int
	maxTotalBends       int
	maxRouteStretch     float64
	maxBridgedCrossings int
	minNodeGap          float64
	minContainerGutter  float64
	minLabelClearance   float64
	minSegmentLength    float64
}

type svgCompositionEdge struct {
	id      string
	route   [][2]float64
	bridges int
}

type svgCompositionNode struct {
	id     string
	bounds geometry.Bounds
}

type svgCompositionContainer struct {
	id     string
	bounds geometry.Bounds
}

type svgCompositionLabel struct {
	id     string
	owner  string
	bounds geometry.Bounds
}

func compositionInput(svgContent string) (svgCompositionContract, []svgCompositionNode, []svgCompositionContainer, []svgCompositionLabel, []svgCompositionEdge, error) {
	contract := svgCompositionContract{profile: "standard", maxBendsPerEdge: 12, maxTotalBends: 100, maxRouteStretch: 5, maxBridgedCrossings: 8, minLabelClearance: 2}
	decoder := xml.NewDecoder(strings.NewReader(svgContent))
	var nodes []svgCompositionNode
	var containers []svgCompositionContainer
	var labels []svgCompositionLabel
	var edges []svgCompositionEdge
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return contract, nodes, containers, labels, edges, nil
		}
		if err != nil {
			return contract, nil, nil, nil, nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := map[string]string{}
		for _, attr := range start.Attr {
			attrs[attr.Name.Local] = attr.Value
		}
		if start.Name.Local == "svg" {
			if value := strings.TrimSpace(attrs["data-quality-profile"]); value != "" {
				contract.profile = strings.ToLower(value)
			}
			if contract.profile == "showcase" {
				contract.maxBendsPerEdge = 2
				contract.maxTotalBends = 8
				contract.maxRouteStretch = 1.35
				contract.maxBridgedCrossings = 0
				contract.minNodeGap = 40
				contract.minContainerGutter = 20
				contract.minLabelClearance = 4
				contract.minSegmentLength = 16
			}
			compositionApplyContractOverrides(&contract, attrs)
		}
		if attrs["data-graph-role"] == "node" {
			if bounds, ok := parseGraphBounds(attrs["data-graph-bounds"]); ok {
				id := attrs["data-node-id"]
				if id == "" {
					id = attrs["id"]
				}
				if id == "" {
					id = "node"
				}
				nodes = append(nodes, svgCompositionNode{id: id, bounds: bounds})
			}
		}
		if attrs["data-graph-role"] == "container" {
			if bounds, ok := parseGraphBounds(attrs["data-graph-bounds"]); ok {
				id := attrs["data-container-id"]
				if id == "" {
					id = attrs["id"]
				}
				if id == "" {
					id = "container"
				}
				containers = append(containers, svgCompositionContainer{id: id, bounds: bounds})
			}
		}
		if attrs["data-graph-role"] == "label" {
			if bounds, ok := parseGraphBounds(attrs["data-graph-bounds"]); ok {
				id := attrs["id"]
				if id == "" {
					id = "label"
				}
				labels = append(labels, svgCompositionLabel{id: id, owner: attrs["data-owner"], bounds: bounds})
			}
		}
		if start.Name.Local != "path" || attrs["data-graph-role"] != "edge" {
			continue
		}
		id := attrs["data-edge-id"]
		if id == "" {
			id = attrs["id"]
		}
		if id == "" {
			id = "edge"
		}
		edges = append(edges, svgCompositionEdge{id: id, route: parseCoords(attrs["d"]), bridges: len(parseSVGPoints(attrs["data-bridges"]))})
	}
}

func parseGraphBounds(raw string) (geometry.Bounds, bool) {
	values := numberPattern.FindAllString(raw, -1)
	if len(values) != 4 {
		return geometry.Bounds{}, false
	}
	parsed := [4]float64{}
	for index, value := range values {
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return geometry.Bounds{}, false
		}
		parsed[index] = number
	}
	return geometry.Bounds{Left: parsed[0], Top: parsed[1], Right: parsed[2], Bottom: parsed[3]}, parsed[2] >= parsed[0] && parsed[3] >= parsed[1]
}

func compositionRectangleGap(first, second geometry.Bounds) float64 {
	horizontal := math.Max(math.Max(first.Left-second.Right, second.Left-first.Right), 0)
	vertical := math.Max(math.Max(first.Top-second.Bottom, second.Top-first.Bottom), 0)
	return math.Hypot(horizontal, vertical)
}

func compositionContainingContainer(node geometry.Bounds, containers []svgCompositionContainer) (svgCompositionContainer, bool) {
	centerX := (node.Left + node.Right) / 2
	centerY := (node.Top + node.Bottom) / 2
	var best svgCompositionContainer
	bestArea := math.Inf(1)
	for _, container := range containers {
		bounds := container.bounds
		if centerX < bounds.Left || centerX > bounds.Right || centerY < bounds.Top || centerY > bounds.Bottom {
			continue
		}
		area := math.Max(0, bounds.Right-bounds.Left) * math.Max(0, bounds.Bottom-bounds.Top)
		if area < bestArea {
			best, bestArea = container, area
		}
	}
	return best, !math.IsInf(bestArea, 1)
}

func compositionContainerGutter(node, container geometry.Bounds) float64 {
	return math.Min(math.Min(node.Left-container.Left, node.Top-container.Top), math.Min(container.Right-node.Right, container.Bottom-node.Bottom))
}

func compositionExpandBounds(bounds geometry.Bounds, padding float64) geometry.Bounds {
	return geometry.Bounds{Left: bounds.Left - padding, Top: bounds.Top - padding, Right: bounds.Right + padding, Bottom: bounds.Bottom + padding}
}

func compositionBoundsIntersect(first, second geometry.Bounds) bool {
	return first.Right > second.Left && second.Right > first.Left && first.Bottom > second.Top && second.Bottom > first.Top
}

func compositionRouteHitsBounds(points [][2]float64, bounds geometry.Bounds) bool {
	for index := 0; index+1 < len(points); index++ {
		start := geometry.Point{X: points[index][0], Y: points[index][1]}
		end := geometry.Point{X: points[index+1][0], Y: points[index+1][1]}
		if segmentHitsInterior(start, end, bounds) {
			return true
		}
	}
	return false
}

func compositionRouteStretch(points [][2]float64) float64 {
	if len(points) < 2 {
		return 1
	}
	length := 0.0
	for index := 0; index+1 < len(points); index++ {
		length += math.Abs(points[index+1][0]-points[index][0]) + math.Abs(points[index+1][1]-points[index][1])
	}
	direct := math.Abs(points[len(points)-1][0]-points[0][0]) + math.Abs(points[len(points)-1][1]-points[0][1])
	if direct <= 1e-9 {
		return 1
	}
	return length / direct
}

func compositionShortestSegment(points [][2]float64) float64 {
	shortest := math.Inf(1)
	for index := 0; index+1 < len(points); index++ {
		length := math.Abs(points[index+1][0]-points[index][0]) + math.Abs(points[index+1][1]-points[index][1])
		if length > 1e-6 {
			shortest = math.Min(shortest, length)
		}
	}
	if math.IsInf(shortest, 1) {
		return -1
	}
	return shortest
}

func compositionApplyContractOverrides(contract *svgCompositionContract, attrs map[string]string) {
	if value, ok := parseSVGNumber(attrs["data-max-route-stretch"]); ok {
		contract.maxRouteStretch = value
	}
	if value, ok := parseSVGNumber(attrs["data-min-node-gap"]); ok {
		contract.minNodeGap = value
	}
	if value, ok := parseSVGNumber(attrs["data-min-container-gutter"]); ok {
		contract.minContainerGutter = value
	}
	if value, ok := parseSVGNumber(attrs["data-min-label-clearance"]); ok {
		contract.minLabelClearance = value
	}
	if value, ok := parseSVGNumber(attrs["data-min-segment-length"]); ok {
		contract.minSegmentLength = value
	}
	if value, ok := parseSVGNumber(attrs["data-max-bends-per-edge"]); ok {
		contract.maxBendsPerEdge = int(value)
	}
	if value, ok := parseSVGNumber(attrs["data-max-total-bends"]); ok {
		contract.maxTotalBends = int(value)
	}
	if value, ok := parseSVGNumber(attrs["data-max-bridged-crossings"]); ok {
		contract.maxBridgedCrossings = int(value)
	}
}

func countBendsPathD(d string) int {
	return countBendsPoints(parseCoords(d))
}

func countBendsPoints(coords [][2]float64) int {
	if len(coords) < 3 {
		return 0
	}
	bends := 0
	lastAxis := ""
	for i := 1; i < len(coords); i++ {
		dx := coords[i][0] - coords[i-1][0]
		dy := coords[i][1] - coords[i-1][1]
		axis := "h"
		if math.Abs(dy) > math.Abs(dx) {
			axis = "v"
		}
		if lastAxis != "" && axis != lastAxis {
			bends++
		}
		lastAxis = axis
	}
	return bends
}

func parseCoords(d string) [][2]float64 {
	var coords [][2]float64
	fields := strings.Fields(d)
	for i := 0; i < len(fields); i++ {
		if fields[i] == "M" || fields[i] == "L" {
			if i+2 < len(fields) {
				var x, y float64
				if n, _ := fmt.Sscanf(fields[i+1]+" "+fields[i+2], "%f %f", &x, &y); n == 2 {
					coords = append(coords, [2]float64{x, y})
				}
				i += 2
			}
		}
	}
	return coords
}

// ValidateSVG runs all checks.
func ValidateSVG(svgContent string) map[string]CheckResult {
	return map[string]CheckResult{
		"xml":         CheckXML(svgContent),
		"markers":     CheckMarkers(svgContent),
		"collisions":  CheckCollisions(svgContent),
		"geometry":    CheckGeometry(svgContent),
		"composition": CheckComposition(svgContent),
	}
}

// AllOk returns true if all validation checks passed.
func AllOk(results map[string]CheckResult) bool {
	for _, r := range results {
		if !r.Ok {
			return false
		}
	}
	return true
}

// --- Geometry helpers (ported from composition_quality.py) ---

// CompositionReport summarizes composition quality metrics.
type CompositionReport struct {
	Crossings     int     `json:"crossings"`
	BridgeJumps   int     `json:"bridge_jumps"`
	MaxBends      int     `json:"max_bends"`
	TotalBends    int     `json:"total_bends"`
	RouteStretch  float64 `json:"route_stretch"`
	MinSpacing    float64 `json:"min_spacing"`
	MinGutter     float64 `json:"min_gutter"`
	ShowcaseReady bool    `json:"showcase_ready"`
}

// Bounds type alias
type Bounds = geometry.Bounds

// EvaluateComposition checks a diagram against the showcase quality contract.
func EvaluateComposition(nodes []Bounds, routes [][]geometry.Point) CompositionReport {
	report := CompositionReport{
		RouteStretch: 1.0,
		MinSpacing:   math.MaxFloat64,
		MinGutter:    math.MaxFloat64,
	}

	// Node-to-node spacing
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			h := math.Max(math.Max(nodes[i].Left-nodes[j].Right, nodes[j].Left-nodes[i].Right), 0)
			v := math.Max(math.Max(nodes[i].Top-nodes[j].Bottom, nodes[j].Top-nodes[i].Bottom), 0)
			spacing := math.Hypot(h, v)
			if spacing < report.MinSpacing {
				report.MinSpacing = spacing
			}
		}
	}

	// Route analysis
	var totalDirect, totalActual float64
	for _, route := range routes {
		length := geometry.RouteLength(route)
		bends := geometry.BendCount(route)
		if bends > report.MaxBends {
			report.MaxBends = bends
		}
		report.TotalBends += bends

		if len(route) >= 2 {
			direct := math.Hypot(route[len(route)-1].X-route[0].X, route[len(route)-1].Y-route[0].Y)
			totalDirect += direct
			totalActual += length
			if direct > 0 {
				stretch := length / direct
				if stretch > report.RouteStretch {
					report.RouteStretch = stretch
				}
			}
		}
	}

	// Check against showcase limits
	report.ShowcaseReady = report.Crossings == 0 &&
		report.BridgeJumps == 0 &&
		report.MaxBends <= 2 &&
		report.TotalBends <= 8 &&
		report.RouteStretch <= 1.35 &&
		report.MinSpacing >= 40 &&
		report.MinGutter >= 20

	return report
}
