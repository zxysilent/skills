package semantics

import (
	"encoding/json"
	"os"
	"testing"
)

func TestValidateOpsNormalizesSignalsAndCriticalPath(t *testing.T) {
	data := map[string]any{
		"diagram_type":       "observability",
		"observation_window": "5m",
		"nodes": []any{
			map[string]any{"id": "gateway", "ops_role": "service", "x": 0.0, "y": 0.0, "width": 180.0, "height": 108.0, "status": "ok", "status_label": "HEALTHY", "signals": map[string]any{
				"latency":    map[string]any{"value": "10", "unit": "ms", "window": "5m", "status": "ok"},
				"traffic":    map[string]any{"value": "1", "unit": "rps", "window": "5m", "status": "ok"},
				"errors":     map[string]any{"value": "0", "unit": "%", "window": "5m", "status": "ok"},
				"saturation": map[string]any{"value": "20", "unit": "%", "window": "5m", "status": "ok"},
			}},
			map[string]any{"id": "checkout", "ops_role": "service", "x": 220.0, "y": 0.0, "width": 180.0, "height": 108.0, "status": "critical", "status_label": "DEGRADED", "signals": map[string]any{
				"latency":    map[string]any{"value": "400", "unit": "ms", "window": "5m", "status": "critical"},
				"traffic":    map[string]any{"value": "1", "unit": "rps", "window": "5m", "status": "ok"},
				"errors":     map[string]any{"value": "3", "unit": "%", "window": "5m", "status": "critical"},
				"saturation": map[string]any{"value": "80", "unit": "%", "window": "5m", "status": "warn"},
			}},
			map[string]any{"id": "root", "ops_role": "trace_span", "span_id": "root", "start_ms": 0.0, "duration_ms": 100.0, "x": 0.0, "y": 180.0, "width": 200.0, "height": 28.0},
		},
		"arrows":        []any{map[string]any{"id": "request", "source": "gateway", "target": "checkout", "edge_kind": "business", "flow": "control"}},
		"critical_path": []any{"request"},
	}

	if err := validateOps(data); err != nil {
		t.Fatalf("validateOps(): %v", err)
	}

	nodes := getNodeMap(data)
	badges, ok := nodes["gateway"]["metric_badges"].([]any)
	if !ok || len(badges) != 4 {
		t.Fatalf("metric_badges = %#v, want four normalized signals", nodes["gateway"]["metric_badges"])
	}
	edge := getEdgeMap(data)["request"]
	if edge["critical"] != true || edge["critical_hop"] != 1 || edge["critical_hops"] != 1 {
		t.Fatalf("critical path annotation = %#v", edge)
	}
}

func TestCloudAndEventFixturesValidateAndEnrich(t *testing.T) {
	for _, test := range []struct {
		name    string
		path    string
		style   int
		profile string
	}{
		{"cloud", "../../fixtures/cloud-fabric-style10.json", 10, "cloud-fabric"},
		{"event", "../../fixtures/event-transit-style11.json", 11, "event-transit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			bytes, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			var data map[string]any
			if err := json.Unmarshal(bytes, &data); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			if _, err := ValidateSemanticContract(test.style, test.profile, data); err != nil {
				t.Fatalf("ValidateSemanticContract(): %v", err)
			}
		})
	}
}

func TestCloudRejectsUnknownIcon(t *testing.T) {
	data := map[string]any{
		"diagram_type": "deployment", "platform_profile": "aws", "icon_manifest_version": cloudManifestVersion,
		"containers": []any{map[string]any{"id": "region", "deployment_kind": "region", "x": 0.0, "y": 0.0, "width": 300.0, "height": 300.0}},
		"nodes":      []any{map[string]any{"id": "service", "deployment_id": "region", "icon_id": "missing", "x": 30.0, "y": 30.0, "width": 100.0, "height": 80.0}},
		"arrows":     []any{},
	}
	if err := validateCloud(data); err == nil || err.(*ContractError).Code != "CLOUD_ICON_UNKNOWN" {
		t.Fatalf("validateCloud() error = %v, want CLOUD_ICON_UNKNOWN", err)
	}
}

func TestOpsRejectsSignalOutsideObservationWindow(t *testing.T) {
	bytes, err := os.ReadFile("../../fixtures/ops-pulse-style12.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(bytes, &data); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	nodes := getNodeMap(data)
	signals := nodes["edge-gateway"]["signals"].(map[string]any)
	signals["latency"].(map[string]any)["window"] = "1m"
	if err := validateOps(data); err == nil || err.(*ContractError).Code != "OPS_OBSERVATION_WINDOW" {
		t.Fatalf("validateOps() error = %v, want OPS_OBSERVATION_WINDOW", err)
	}
}
