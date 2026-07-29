package render

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"drawcli/ir"
	"drawcli/semantics"
	"drawcli/validate"
)

func TestOpsPulseFixtureRendersOriginalSpecializedElements(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/ops-pulse-style12.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if _, err := semantics.ValidateSemanticContract(12, "ops-pulse", source); err != nil {
		t.Fatalf("normalize Ops Pulse semantics: %v", err)
	}
	diagram, err := ir.Normalize(source)
	if err != nil {
		t.Fatalf("normalize fixture: %v", err)
	}
	diagram.StyleIndex = 12

	svg := RenderSVG(diagram, GetProfile(12))
	for _, want := range []string{
		"@5m",
		">0%</text",
		">25%</text",
		">50%</text",
		">75%</text",
		">100%</text",
		`fill="#38bdf8" fill-opacity="0.12"`,
		">CRITICAL · CORRELATED TRACE</text",
		`id="edge-gateway-api-hop"`,
		">1/3</text",
		`d="M 220 224 L 270 224"`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("RenderSVG() missing original Ops Pulse element %q", want)
		}
	}
}

func TestBlueprintFixtureHonorsContainerAndTitleBlockData(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/microservices-style3.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	diagram, err := ir.Normalize(source)
	if err != nil {
		t.Fatalf("normalize fixture: %v", err)
	}
	diagram.StyleIndex = 3
	svg := RenderSVG(diagram, GetProfile(3))
	for _, want := range []string{
		">01 // EDGE</text",
		`<rect x="700" y="620" width="220" height="76"`,
		">AI MICROSERVICES</text",
		">REV: 1.1</text",
		`<ellipse cx="150" cy="486" rx="35"`,
		`stroke-opacity="0.45"`,
		`<g id="container-edge" data-graph-role="container" data-container-id="edge"`,
		`data-graph-bounds="40,110,920,220"`,
		`<g id="node-postgres" data-graph-role="node" data-node-id="postgres"`,
		`data-graph-bounds="80,480,220,530"`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("RenderSVG() missing original Blueprint element %q", want)
		}
	}
}

func TestGlassmorphismFixtureDistributesParallelPortsLikeOriginal(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/multi-agent-style5.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	diagram, err := ir.Normalize(source)
	if err != nil {
		t.Fatalf("normalize fixture: %v", err)
	}
	diagram.StyleIndex = 5
	svg := RenderSVG(diagram, GetProfile(5))
	for _, want := range []string{
		`d="M 472 200 L 472 306 L 170 306 L 170 330"`,
		`d="M 490 200 L 490 330"`,
		`d="M 508 200 L 508 220 L 790 220 L 790 330"`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("RenderSVG() missing original Glassmorphism route %q", want)
		}
	}
}

func TestRenderSVGEmbedsPortableCompositionContract(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/mem0-style1.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	diagram, err := ir.Normalize(source)
	if err != nil {
		t.Fatalf("normalize fixture: %v", err)
	}

	svg := RenderSVG(diagram, GetProfile(1))
	for _, want := range []string{
		`data-quality-profile="showcase"`,
		`data-max-bends-per-edge="2"`,
		`data-max-total-bends="8"`,
		`data-max-route-stretch="1.35"`,
		`data-max-bridged-crossings="0"`,
		`data-min-node-gap="40"`,
		`data-min-container-gutter="20"`,
		`data-min-label-clearance="4"`,
		`data-min-segment-length="16"`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("RenderSVG() missing portable composition contract attribute %q", want)
		}
	}
}

func TestRenderSVGUsesValidExplicitWaypointsWithoutExtraDetours(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/quality-baseline/agent-runtime-style1.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	diagram, err := ir.Normalize(source)
	if err != nil {
		t.Fatalf("normalize fixture: %v", err)
	}

	svg := RenderSVG(diagram, GetProfile(1))
	for _, want := range []string{
		`id="context" data-graph-role="edge" data-edge-id="context"`,
		`id="z-feedback" data-graph-role="edge" data-edge-id="z-feedback"`,
		`d="M 632 216 L 632 270 L 390 270 L 390 344"`,
		`d="M 845 344 L 845 270 L 668 270 L 668 216"`,
		`<g id="context-label" data-graph-role="label" data-owner="context" data-graph-bounds="`,
		`<g id="trace-label" data-graph-role="label" data-owner="trace" data-graph-bounds="740.50,320.00,789.50,340.00">`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("RenderSVG() did not preserve a valid explicit route %q", want)
		}
	}
}

func TestRenderSVGKeepsLabelsClearOfExplicitRoutes(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/quality-baseline/agent-runtime-style1.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	diagram, err := ir.Normalize(source)
	if err != nil {
		t.Fatalf("normalize fixture: %v", err)
	}

	result := validate.CheckComposition(RenderSVG(diagram, GetProfile(1)))
	if !result.Ok {
		t.Fatalf("RenderSVG() placed a label inside the composition clearance: %#v", result.Details)
	}
}

func TestRenderSVGAllocatesSafeDOMIDsAndPreservesSemanticAttributes(t *testing.T) {
	data := map[string]any{
		"mode":         "architecture",
		"width":        480,
		"height":       240,
		"style":        12,
		"motion_scene": "request-trace",
		"nodes": []any{
			map[string]any{"id": "a b", "x": 20, "y": 80, "width": 80, "height": 50, "motion_role": "producer", "motion_stage": 1, "motion_order": 2, "status": "healthy", "span_id": "span-a"},
			map[string]any{"id": "a-b", "x": 340, "y": 80, "width": 80, "height": 50, "motion_role": "consumer", "motion_stage": 3, "motion_order": 4, "parent_span": "span-a", "duration_ms": 12},
			map[string]any{"id": "中文", "x": 190, "y": 160, "width": 80, "height": 50},
		},
		"arrows": []any{
			map[string]any{"id": "edge one", "source": "a b", "target": "a-b", "flow": "data", "edge_kind": "request", "topic_id": "orders", "protocol": "grpc", "via": "gateway", "critical": true, "critical_path_id": "p1", "critical_hop": 1, "critical_hops": 1, "motion_role": "request", "motion_stage": 2, "motion_order": 3},
			map[string]any{"id": "edge-one", "source": "a-b", "target": "a b", "flow": "feedback"},
		},
	}

	diagram, err := ir.Normalize(data)
	if err != nil {
		t.Fatalf("normalize diagram: %v", err)
	}
	diagram.StyleIndex = 12
	svg := RenderSVG(diagram, GetProfile(12))

	for _, want := range []string{
		`id="node-a-b"`,
		`id="node-a-b-2"`,
		`id="node-node"`,
		`id="edge-one"`,
		`id="edge-one-2"`,
		`data-motion-scene="request-trace"`,
		`data-node-id="a b"`,
		`data-motion-role="producer"`,
		`data-status="healthy"`,
		`data-edge-id="edge one"`,
		`data-source="a b"`,
		`data-target="a-b"`,
		`data-edge-kind="request"`,
		`data-topic-id="orders"`,
		`data-flow="data"`,
		`data-protocol="grpc"`,
		`data-via="gateway"`,
		`data-critical-path-id="p1"`,
		`data-critical="true"`,
		`data-bends="`,
		`data-route-stretch="`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("RenderSVG() missing %q", want)
		}
	}
}
