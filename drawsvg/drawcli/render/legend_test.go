package render

import (
	"strings"
	"testing"

	"drawcli/ir"
)

func TestLegendEntryAdvanceFitsLongLabels(t *testing.T) {
	for _, label := range []string{"memory write", "memory read", "data transform"} {
		if got := legendEntryAdvance(label); got <= 100 {
			t.Fatalf("legend entry %q advance = %.1f, want more than fixed 100px slot", label, got)
		}
	}
}

func TestUserAvatarIncludesOuterNodeShape(t *testing.T) {
	diagram, err := ir.Normalize(map[string]any{
		"mode": "memory", "style": 1, "width": 240, "height": 160,
		"nodes": []any{map[string]any{
			"id": "user", "kind": "user_avatar", "x": 60, "y": 40,
			"width": 120, "height": 54, "label": "User",
			"fill": "#eff6ff", "stroke": "#3b82f6", "flat": true,
		}},
	})
	if err != nil {
		t.Fatalf("normalize diagram: %v", err)
	}
	svg := RenderSVG(diagram, GetProfile(1))
	if !strings.Contains(svg, `<rect x="60" y="40" width="120" height="54" rx="10" fill="#eff6ff" stroke="#3b82f6" stroke-width="2.0"/>`) {
		t.Fatal("user_avatar should include the rounded outer node shape")
	}
	if !strings.Contains(svg, `<text x="124" y="73" class="node-title" font-size="18.0">User</text>`) {
		t.Fatal("user_avatar label should align with the original avatar layout")
	}
}

func TestStandardNodeTitleUsesFittedSizeAndBaseline(t *testing.T) {
	diagram, err := ir.Normalize(map[string]any{
		"mode": "memory", "style": 1, "width": 240, "height": 160,
		"nodes": []any{map[string]any{
			"id": "app", "kind": "rect", "x": 50, "y": 40,
			"width": 160, "height": 54, "label": "AI App / Agent",
			"sublabel": "conversation input", "fill": "#fff", "stroke": "#ccc",
		}},
	})
	if err != nil {
		t.Fatalf("normalize diagram: %v", err)
	}
	svg := RenderSVG(diagram, GetProfile(1))
	if !strings.Contains(svg, `<text x="130" y="67" class="node-title" text-anchor="middle" font-size="17.48">AI App / Agent</text>`) {
		t.Fatal("standard node title should use fitted font size and centered baseline")
	}
}

func TestEventTransitSignatureKeepsMarkersInsideBadge(t *testing.T) {
	diagram, err := ir.Normalize(map[string]any{
		"mode": "memory", "style": 11, "width": 960, "height": 160,
	})
	if err != nil {
		t.Fatalf("normalize diagram: %v", err)
	}
	diagram.StyleIndex = 11
	svg := RenderSVG(diagram, GetProfile(11))
	if !strings.Contains(svg, `y1="39"`) || !strings.Contains(svg, `y2="39"`) || !strings.Contains(svg, `cy="39"`) {
		t.Fatal("event transit signature markers should use the badge-relative y position")
	}
}

func TestCloudServiceUsesSemanticGlyphAndColor(t *testing.T) {
	diagram, err := ir.Normalize(map[string]any{
		"mode": "architecture", "style": 10, "width": 240, "height": 160,
		"nodes": []any{map[string]any{
			"id": "queue", "kind": "cloud_service", "x": 30, "y": 40,
			"width": 180, "height": 72, "label": "Work Queue", "sublabel": "async jobs",
			"glyph": "queue", "icon_color": "#d97706", "provider": "GENERIC",
		}},
	})
	if err != nil {
		t.Fatalf("normalize diagram: %v", err)
	}
	svg := RenderSVG(diagram, GetProfile(10))
	if !strings.Contains(svg, `data-cloud-glyph="queue"`) {
		t.Fatal("cloud service should render the semantic queue glyph")
	}
	if !strings.Contains(svg, `stroke="#d97706"`) {
		t.Fatal("cloud service should render with the semantic icon color")
	}
}
