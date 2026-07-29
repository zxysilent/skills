package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderRejectsAIAuthoredStyleEight(t *testing.T) {
	input := filepath.Join(t.TempDir(), "style-8.json")
	if err := os.WriteFile(input, []byte(`{
  "mode": "architecture",
  "style": 8,
  "nodes": [{"id":"source","x":20,"y":40,"width":80,"height":40}],
  "arrows": []
}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	err := cmdRender([]string{"architecture", input, filepath.Join(t.TempDir(), "out.svg")})
	if err == nil || !strings.Contains(err.Error(), "AI-authored") {
		t.Fatalf("cmdRender() error = %v, want AI-authored style rejection", err)
	}
}

func TestCheckHonorsSelectedChecks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagram.svg")
	if err := os.WriteFile(path, []byte(`<svg viewBox="0 0 20 20"/>`), 0o644); err != nil {
		t.Fatalf("write SVG: %v", err)
	}

	output := captureStdout(t, func() error {
		return cmdCheck([]string{path, "--check", "xml"})
	})
	if !strings.Contains(output, `"xml"`) || strings.Contains(output, `"markers"`) || strings.Contains(output, `"collisions"`) {
		t.Fatalf("cmdCheck() output = %s, want only the requested XML check", output)
	}
}

func TestInspectMatchesSVGSemanticMetadataContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagram.svg")
	svg := `<svg viewBox="0 0 20 20" data-generator="drawcli-go" data-schema-version="1" data-style-id="12" data-visual-theme="Ops Pulse" data-diagram-type="architecture" data-semantic-profile="ops-pulse" data-semantic-valid="true"><g data-graph-role="node"/><path data-graph-role="edge"/></svg>`
	if err := os.WriteFile(path, []byte(svg), 0o644); err != nil {
		t.Fatalf("write SVG: %v", err)
	}

	output := captureStdout(t, func() error { return cmdInspect([]string{path}) })
	for _, want := range []string{
		`"generator": "drawcli-go"`,
		`"schema_version": "1"`,
		`"style_id": "12"`,
		`"semantic_profile": "ops-pulse"`,
		`"semantic_valid": "true"`,
		`"roles": {`,
		`"node": 1`,
		`"edge": 1`,
	} {
		if !strings.Contains(output, want) {
			t.Errorf("cmdInspect() missing %q in %s", want, output)
		}
	}
}

func TestRenderReportContainsPythonContractFields(t *testing.T) {
	input, err := filepath.Abs("../fixtures/mem0-style1.json")
	if err != nil {
		t.Fatalf("fixture path: %v", err)
	}
	dir := t.TempDir()
	report := filepath.Join(dir, "report.json")
	captureStdout(t, func() error {
		return cmdRender([]string{"memory", input, filepath.Join(dir, "diagram.svg"), "--report", report})
	})
	contents, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(contents, &decoded); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	for _, key := range []string{"schema_version", "input_schema", "mode", "style", "semantics", "canvas", "text_metrics", "placements", "edges", "composition", "issues", "summary"} {
		if _, found := decoded[key]; !found {
			t.Errorf("render report missing %q: %s", key, contents)
		}
	}
	if edges, ok := decoded["edges"].([]any); !ok || len(edges) != 8 {
		t.Errorf("render report edges = %#v, want eight detailed routes", decoded["edges"])
	}
}

func captureStdout(t *testing.T, action func() error) string {
	t.Helper()
	previous := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("open stdout pipe: %v", err)
	}
	os.Stdout = writer
	err = action()
	writer.Close()
	os.Stdout = previous
	if err != nil {
		t.Fatalf("action: %v", err)
	}
	output, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(output)
}
