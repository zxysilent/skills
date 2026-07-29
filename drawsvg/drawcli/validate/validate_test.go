package validate

import (
	"strings"
	"testing"
)

func TestCheckGeometryRejectsEdgeCrossingNode(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <defs><marker id="arrow-main"><path d="M0 0 L8 4 L0 8 Z"/></marker></defs>
  <rect id="blocker" data-graph-role="node" x="160" y="80" width="80" height="60"/>
  <path id="edge" data-graph-role="edge" d="M 20 110 H 280" marker-end="url(#arrow-main)"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted an edge crossing a node: %#v", result.Details)
	}
}

func TestCheckGeometryAppliesParentTransforms(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <g transform="translate(100 0)">
    <rect id="blocker" data-graph-role="node" x="100" y="80" width="80" height="60"/>
  </g>
  <path id="edge" data-graph-role="edge" d="M 0 110 H 180"/>
</svg>`

	result := CheckGeometry(svg)
	if !result.Ok {
		t.Fatalf("CheckGeometry() ignored a parent transform and reported a false collision: %#v", result.Details)
	}
}

func TestCheckGeometrySamplesQuadraticCurves(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <rect id="blocker" data-graph-role="node" x="180" y="85" width="40" height="30"/>
  <path id="edge" data-graph-role="edge" d="M 20 20 Q 200 200 380 20"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted a quadratic edge crossing a node: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "intersects") {
		t.Fatalf("CheckGeometry() did not report a geometric curve collision: %#v", result.Details)
	}
}

func TestCheckGeometrySamplesArcs(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <rect id="blocker" data-graph-role="node" x="180" y="0" width="40" height="30"/>
  <path id="edge" data-graph-role="edge" d="M 20 100 A 180 100 0 0 1 380 100"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted an arc edge crossing a node: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "intersects") {
		t.Fatalf("CheckGeometry() did not report a geometric arc collision: %#v", result.Details)
	}
}

func TestCheckGeometrySamplesSmoothCubicCurves(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 260">
	  <rect id="blocker" data-graph-role="node" x="270" y="190" width="40" height="30"/>
  <path id="edge" data-graph-role="edge" d="M 20 120 C 80 0 120 0 200 120 S 320 240 380 120"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted a smooth cubic edge crossing a node: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "intersects") {
		t.Fatalf("CheckGeometry() did not sample the smooth cubic curve: %#v", result.Details)
	}
}

func TestCheckGeometrySamplesSmoothQuadraticCurves(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 260">
	  <rect id="blocker" data-graph-role="node" x="270" y="165" width="40" height="30"/>
  <path id="edge" data-graph-role="edge" d="M 20 120 Q 100 0 200 120 T 380 120"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted a smooth quadratic edge crossing a node: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "intersects") {
		t.Fatalf("CheckGeometry() did not sample the smooth quadratic curve: %#v", result.Details)
	}
}

func TestCheckGeometryInheritsNodeRoleFromGroup(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <g data-graph-role="node"><rect id="blocker" x="160" y="80" width="80" height="60"/></g>
  <path id="edge" data-graph-role="edge" d="M 20 110 H 280"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() did not inherit data-graph-role=node from the parent group: %#v", result.Details)
	}
}

func TestCheckGeometryDetectsPolygonNodes(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <g data-graph-role="node"><polygon id="blocker" points="160,80 240,80 260,110 240,140 160,140 140,110"/></g>
  <path id="edge" data-graph-role="edge" d="M 20 110 H 280"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted an edge crossing a polygon node: %#v", result.Details)
	}
}

func TestCheckGeometryDetectsPathNodes(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <g data-graph-role="node"><path id="blocker" d="M 160 80 H 240 V 140 H 160 Z"/></g>
  <path id="edge" data-graph-role="edge" d="M 20 110 H 280"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted an edge crossing a path node: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "intersects") {
		t.Fatalf("CheckGeometry() did not report a geometric path collision: %#v", result.Details)
	}
}

func TestCheckGeometryDoesNotTreatLabelAsNode(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <path id="edge" data-graph-role="edge" d="M 20 110 H 380"/>
  <g id="edge-label" data-graph-role="label" data-owner="edge" data-graph-bounds="180,100,220,120"><rect x="180" y="100" width="40" height="20"/></g>
</svg>`

	result := CheckGeometry(svg)
	if !result.Ok {
		t.Fatalf("CheckGeometry() treated a label as a node obstacle: %#v", result.Details)
	}
}

func TestCheckGeometryRejectsOverlappingEdges(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240">
  <path id="first" data-graph-role="edge" d="M 20 110 L 220 110"/>
  <path id="second" data-graph-role="edge" d="M 120 110 L 380 110"/>
</svg>`

	result := CheckGeometry(svg)
	if result.Ok {
		t.Fatalf("CheckGeometry() accepted overlapping edges: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "edge_overlap") {
		t.Fatalf("CheckGeometry() did not report edge overlap: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsShowcaseRouteStretch(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240"
  data-quality-profile="showcase" data-max-route-stretch="1.35">
  <path id="detour" data-graph-role="edge" data-edge-id="detour" d="M 20 20 L 20 180 L 380 180 L 380 20"/>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted an over-stretched showcase route: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_edge_route_stretch") {
		t.Fatalf("CheckComposition() did not report route stretch: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsShowcaseNodeGap(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-quality-profile="showcase">
  <g id="node-left" data-graph-role="node" data-node-id="left" data-graph-bounds="20,40,120,100"><rect x="20" y="40" width="100" height="60"/></g>
  <g id="node-right" data-graph-role="node" data-node-id="right" data-graph-bounds="130,40,230,100"><rect x="130" y="40" width="100" height="60"/></g>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted nodes closer than the showcase gap: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_node_gap") {
		t.Fatalf("CheckComposition() did not report node gap: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsShowcaseContainerGutter(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-quality-profile="showcase">
  <g id="service-lane" data-graph-role="container" data-graph-bounds="20,20,380,220"><rect x="20" y="20" width="360" height="200"/></g>
  <g id="node-api" data-graph-role="node" data-node-id="api" data-graph-bounds="25,80,125,140"><rect x="25" y="80" width="100" height="60"/></g>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted insufficient container gutter: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_container_gutter") {
		t.Fatalf("CheckComposition() did not report container gutter: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsShowcaseMicroSegment(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-quality-profile="showcase">
  <path id="tiny-turn" data-graph-role="edge" data-edge-id="tiny-turn" d="M 20 40 L 30 40 L 30 160"/>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted a showcase micro segment: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_edge_micro_segment") {
		t.Fatalf("CheckComposition() did not report the micro segment: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsConfiguredBendBudget(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-max-bends-per-edge="0" data-max-total-bends="0">
  <path id="turn" data-graph-role="edge" data-edge-id="turn" d="M 20 40 L 120 40 L 120 160"/>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted an edge above the configured bend budget: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) < 2 || !strings.Contains(strings.Join(issues, "\n"), "composition_edge_bend_budget") || !strings.Contains(strings.Join(issues, "\n"), "composition_total_bend_budget") {
		t.Fatalf("CheckComposition() did not report bend budget violations: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsShowcaseLabelClearance(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-quality-profile="showcase">
  <g id="node-api" data-graph-role="node" data-node-id="api" data-graph-bounds="100,80,200,140"><rect x="100" y="80" width="100" height="60"/></g>
  <g id="edge-label" data-graph-role="label" data-owner="edge" data-graph-bounds="196,100,246,120"><text x="221" y="110">edge</text></g>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted a label inside showcase clearance: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_label_clearance") {
		t.Fatalf("CheckComposition() did not report label clearance: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsLabelClearanceFromOtherEdge(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-quality-profile="showcase">
  <path id="owner" data-graph-role="edge" data-edge-id="owner" d="M 20 30 L 80 30"/>
  <path id="other" data-graph-role="edge" data-edge-id="other" d="M 20 110 L 380 110"/>
  <g id="owner-label" data-graph-role="label" data-owner="owner" data-graph-bounds="180,100,220,120"><text x="200" y="110">owner</text></g>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted a label overlapping another edge: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_label_clearance") {
		t.Fatalf("CheckComposition() did not report label-to-edge clearance: %#v", result.Details)
	}
}

func TestCheckCompositionRejectsConfiguredBridgeBudget(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 240" data-max-bridged-crossings="0">
  <path id="bridged" data-graph-role="edge" data-edge-id="bridged" data-bridges="100,100" d="M 20 100 L 380 100"/>
</svg>`

	result := CheckComposition(svg)
	if result.Ok {
		t.Fatalf("CheckComposition() accepted a bridge above the configured budget: %#v", result.Details)
	}
	issues, _ := result.Details["issues"].([]string)
	if len(issues) == 0 || !strings.Contains(issues[0], "composition_bridge_budget") {
		t.Fatalf("CheckComposition() did not report bridge budget: %#v", result.Details)
	}
}

func TestValidateSVGReportsDedicatedCollisionCheck(t *testing.T) {
	svg := `<svg viewBox="0 0 200 100">
  <g id="node" data-graph-role="node" data-graph-bounds="80,20,120,80"><rect x="80" y="20" width="40" height="60"/></g>
  <path id="edge" data-graph-role="edge" d="M 10 50 L 190 50"/>
</svg>`

	result, found := ValidateSVG(svg)["collisions"]
	if !found {
		t.Fatal("ValidateSVG() did not include the collisions check")
	}
	if result.Ok {
		t.Fatalf("ValidateSVG() collisions result = %#v, want collision failure", result)
	}
}
