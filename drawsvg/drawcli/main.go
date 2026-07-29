// Command drawcli is the SVG-only diagram generator CLI.
// Ported from drawcli.py (157 lines).
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"drawcli/ir"
	"drawcli/render"
	"drawcli/semantics"
	"drawcli/validate"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "doctor":
		err = cmdDoctor()
	case "validate":
		err = cmdValidate(args)
	case "render":
		err = cmdRender(args)
	case "check":
		err = cmdCheck(args)
	case "inspect":
		err = cmdInspect(args)
	case "examples":
		err = cmdExamples()
	case "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		if _, reported := err.(reportedCheckFailure); reported {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: drawcli <command> [args]

Commands:
  doctor              Check runtime availability
  validate <mode> <json>   Validate and normalize diagram JSON
  render   <mode> <json> <output.svg> [--report report.json]
  check    <svg> [--check xml,markers,geometry,composition]
  inspect  <svg>
  examples            List available fixture examples
`)
}

// --- doctor ---

func cmdDoctor() error {
	fmt.Printf("{\"go\":{\"ok\":true,\"value\":%q},\"generator\":{\"ok\":true}}\n", runtime.Version())
	return nil
}

// --- validate ---

func cmdValidate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: drawcli validate <mode> <input.json>")
	}
	mode := args[0]
	path := args[1]

	data, err := readJSON(path)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	diagram, err := ir.Normalize(data)
	if err != nil {
		return fmt.Errorf("normalize: %w", err)
	}

	styleID, err := semantics.ResolveStyleIndex(data)
	if err != nil {
		return fmt.Errorf("style: %w", err)
	}

	profileHint := ""
	if p, ok := data["semantic_profile"].(string); ok {
		profileHint = p
	}
	report, err := semantics.ValidateSemanticContract(styleID, profileHint, data)
	if err != nil {
		return fmt.Errorf("contract: %w", err)
	}

	out := map[string]any{
		"ok":             true,
		"schema_version": diagram.SchemaVersion,
		"input_schema":   diagram.InputSchema,
		"mode":           diagram.Mode,
		"style": map[string]any{
			"id":   report.Style,
			"name": report.VisualTheme,
		},
		"semantics": report,
		"nodes":     len(diagram.Nodes),
		"edges":     len(diagram.Edges),
	}

	if mode != diagram.Mode && diagram.Mode != "" {
		fmt.Fprintf(os.Stderr, "warning: mode %q differs from fixture mode %q\n", mode, diagram.Mode)
	}

	return printJSON(out)
}

// --- render ---

func cmdRender(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: drawcli render <mode> <input.json> <output.svg> [--report report.json]")
	}
	mode := args[0]
	inputPath := args[1]
	outputPath := args[2]

	reportPath := ""
	for i := 3; i < len(args); i++ {
		if args[i] == "--report" && i+1 < len(args) {
			reportPath = args[i+1]
		}
	}

	data, err := readJSON(inputPath)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	// Resolve style
	styleID, err := semantics.ResolveStyleIndex(data)
	if err != nil {
		return fmt.Errorf("style: %w", err)
	}
	if styleID == 8 {
		return fmt.Errorf("Style 8 (Dark Luxury) is AI-authored and cannot be rendered from a template; hand-craft the SVG using references/style-8-dark-luxury.md")
	}

	// Resolve semantic contract (non-fatal if generic)
	profileHint := ""
	if p, ok := data["semantic_profile"].(string); ok {
		profileHint = p
	}
	semReport, err := semantics.ValidateSemanticContract(styleID, profileHint, data)
	if err != nil {
		return fmt.Errorf("contract: %w", err)
	}

	// Semantic contracts enrich the raw fixture (for example, Ops Pulse marks
	// critical hops). Normalize only after that enrichment reaches the IR.
	diagram, err := ir.Normalize(data)
	if err != nil {
		return fmt.Errorf("normalize: %w", err)
	}
	diagram.StyleIndex = styleID

	// Use fixture's mode if different
	if diagram.Mode == "" {
		diagram.Mode = mode
	}

	if semReport != nil {
		diagram.SemanticProfile = semReport.Profile
		diagram.VisualTheme = semReport.VisualTheme
	} else {
		diagram.SemanticProfile = "generic"
		if name, ok := semantics.StyleNames[styleID]; ok {
			diagram.VisualTheme = name
		}
	}

	profile := render.GetProfile(styleID)
	svg := render.RenderSVG(diagram, profile)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(svg), 0644); err != nil {
		return fmt.Errorf("write SVG: %w", err)
	}

	result := map[string]any{
		"ok":     true,
		"svg":    outputPath,
		"report": nil,
	}
	if reportPath != "" {
		rpt := render.BuildRenderReport(diagram, profile)
		if semReport != nil {
			rpt["semantics"] = semReport
		}
		rptJSON, _ := json.MarshalIndent(rpt, "", "  ")
		if err := os.WriteFile(reportPath, rptJSON, 0644); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
		result["report"] = reportPath
	}

	return printJSON(result)
}

// --- check ---

type reportedCheckFailure struct{}

func (reportedCheckFailure) Error() string { return "SVG checks failed" }

func cmdCheck(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: drawcli check <svg> [--check xml,markers,geometry,composition]")
	}
	svgPath := args[0]
	svgBytes, err := os.ReadFile(svgPath)
	if err != nil {
		return fmt.Errorf("read SVG: %w", err)
	}
	svgContent := string(svgBytes)

	allResults := validate.ValidateSVG(svgContent)
	results, err := selectCheckResults(args[1:], allResults)
	if err != nil {
		return err
	}
	allOK := validate.AllOk(results)

	out := map[string]any{
		"ok":     allOK,
		"checks": results,
	}
	if err := printJSON(out); err != nil {
		return err
	}
	if !allOK {
		return reportedCheckFailure{}
	}
	return nil
}

func selectCheckResults(args []string, available map[string]validate.CheckResult) (map[string]validate.CheckResult, error) {
	selected := make([]string, 0)
	for index := 0; index < len(args); index++ {
		if args[index] != "--check" {
			return nil, fmt.Errorf("unknown check option: %s", args[index])
		}
		if index+1 == len(args) {
			return nil, fmt.Errorf("--check requires a value")
		}
		index++
		selected = append(selected, strings.Split(args[index], ",")...)
	}
	if len(selected) == 0 {
		return available, nil
	}
	results := make(map[string]validate.CheckResult, len(selected))
	for _, name := range selected {
		name = strings.TrimSpace(name)
		result, found := available[name]
		if !found {
			return nil, fmt.Errorf("unknown check: %s", name)
		}
		results[name] = result
	}
	return results, nil
}

// --- inspect ---

func cmdInspect(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: drawcli inspect <svg>")
	}
	svgBytes, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read SVG: %w", err)
	}
	meta, err := inspectSVG(svgBytes)
	if err != nil {
		return fmt.Errorf("inspect SVG: %w", err)
	}
	return printJSON(meta)
}

func inspectSVG(svgBytes []byte) (map[string]any, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(svgBytes)))
	var root map[string]string
	roles := make(map[string]int)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := make(map[string]string, len(start.Attr))
		for _, attr := range start.Attr {
			attrs[attr.Name.Local] = attr.Value
		}
		if root == nil {
			root = attrs
		}
		if role := attrs["data-graph-role"]; role != "" {
			roles[role]++
		}
	}
	if root == nil {
		return nil, fmt.Errorf("SVG has no root element")
	}
	return map[string]any{
		"generator":        root["data-generator"],
		"schema_version":   root["data-schema-version"],
		"style_id":         root["data-style-id"],
		"visual_theme":     root["data-visual-theme"],
		"diagram_type":     root["data-diagram-type"],
		"semantic_profile": root["data-semantic-profile"],
		"semantic_valid":   root["data-semantic-valid"],
		"viewBox":          root["viewBox"],
		"roles":            roles,
	}, nil
}

// --- examples ---

func cmdExamples() error {
	// Look for fixture files relative to the binary location
	candidates := []string{"./fixtures", "../fixtures", "../../fixtures"}
	var found []string
	for _, dir := range candidates {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".json") || strings.HasSuffix(e.Name(), ".svg")) {
				found = append(found, filepath.Join(dir, e.Name()))
			}
		}
		break
	}
	sort.Strings(found)
	return printJSON(map[string]any{"examples": found})
}

// --- helpers ---

func readJSON(path string) (map[string]any, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
