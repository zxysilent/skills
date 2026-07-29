package route

import (
	"drawcli/geometry"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type nodeInfo struct{ x, y, w, h float64 }

func TestFixtureRoute(t *testing.T) {
	b, err := os.ReadFile("../../fixtures/microservices-style3.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	edges, ok := data["arrows"].([]any)
	if !ok {
		t.Fatal("fixture arrows must be an array")
	}
	nodesRaw, ok := data["nodes"].([]any)
	if !ok {
		t.Fatal("fixture nodes must be an array")
	}

	nm := map[string]nodeInfo{}
	var obs []Bounds

	for _, r := range nodesRaw {
		m := r.(map[string]any)
		id := fmt.Sprint(m["id"])
		x, y, w, h := toF(m["x"]), toF(m["y"]), toF(m["width"]), toF(m["height"])
		nm[id] = nodeInfo{x, y, w, h}
		obs = append(obs, Bounds{x, y, x + w, y + h})
	}
	canvas := Bounds{0, 0, 960, 720}
	fail := 0

	for _, r := range edges {
		m := r.(map[string]any)
		src := fmt.Sprint(m["source"])
		tgt := fmt.Sprint(m["target"])
		sp := fmt.Sprint(m["source_port"])
		tp := fmt.Sprint(m["target_port"])

		sn, tn := nm[src], nm[tgt]
		sx, sy := anchor(sn, sp)
		tx, ty := anchor(tn, tp)

		result, err := OrthogonalRoute(
			geometry.Point{X: sx, Y: sy}, geometry.Point{X: tx, Y: ty},
			obs, sp, tp, nil, nil, nil, 24, 18, &canvas, nil,
		)
		if err != nil {
			fmt.Printf("FAIL %s->%s: %v\n", src, tgt, err)
			fail++
		} else {
			fmt.Printf("OK   %s->%s: %d pts %d bends\n", src, tgt, len(result), countBends(result))
		}
	}
	if fail > 0 {
		t.Fatalf("%d/%d failed", fail, len(edges))
	}
}

func anchor(n nodeInfo, p string) (float64, float64) {
	switch p {
	case "right":
		return n.x + n.w, n.y + n.h/2
	case "left":
		return n.x, n.y + n.h/2
	case "bottom":
		return n.x + n.w/2, n.y + n.h
	case "top":
		return n.x + n.w/2, n.y
	default:
		return n.x + n.w/2, n.y + n.h/2
	}
}

func toF(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	default:
		return 0
	}
}

func countBends(pts []geometry.Point) int {
	if len(pts) < 3 {
		return 0
	}
	bends := 0
	last := ""
	for i := 1; i < len(pts); i++ {
		dx := pts[i].X - pts[i-1].X
		dy := pts[i].Y - pts[i-1].Y
		axis := "h"
		if dy*dy > dx*dx {
			axis = "v"
		}
		if last != "" && axis != last {
			bends++
		}
		last = axis
	}
	return bends
}
