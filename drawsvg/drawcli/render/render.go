// Package render generates SVG diagrams from normalized IR.
// Ported from generate-from-template.py (3369 lines).
package render

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"drawcli/geometry"
	"drawcli/ir"
	"drawcli/route"
)

// --- Style Profile ---

type StyleProfile struct {
	Name             string
	FontFamily       string
	Background       string
	Shadow           bool
	TitleAlign       string
	TitleFill        string
	TitleSize        float64
	SubtitleFill     string
	SubtitleSize     float64
	NodeFill         string
	NodeStroke       string
	NodeRadius       float64
	NodeShadow       string
	SectionFill      string
	SectionStroke    string
	SectionDash      string
	SectionLabelFill string
	SectionSubFill   string
	TitleDivider     bool
	SectionUpper     bool
	ArrowWidth       float64
	ArrowColors      map[string]string
	ArrowLabelBG     string
	ArrowLabelOpac   float64
	ArrowLabelFill   string
	TypeLabelFill    string
	TypeLabelSize    float64
	TextPrimary      string
	TextSecondary    string
	TextMuted        string
	LegendFill       string
}

var StyleProfiles = map[int]*StyleProfile{
	1: {
		Name: "Flat Icon", FontFamily: "'Helvetica Neue', Helvetica, Arial, 'PingFang SC', 'Microsoft YaHei', sans-serif",
		Background: "#ffffff", Shadow: true, TitleAlign: "center", TitleFill: "#111827", TitleSize: 30,
		SubtitleFill: "#6b7280", SubtitleSize: 14,
		NodeFill: "#ffffff", NodeStroke: "#d1d5db", NodeRadius: 10, NodeShadow: "url(#shadowSoft)",
		SectionFill: "none", SectionStroke: "#dbe5f1", SectionDash: "6 5",
		SectionLabelFill: "#2563eb", SectionSubFill: "#94a3b8", SectionUpper: true,
		ArrowWidth: 2.4, ArrowLabelBG: "#ffffff", ArrowLabelOpac: 0.94, ArrowLabelFill: "#6b7280",
		TypeLabelFill: "#9ca3af", TypeLabelSize: 12, TextPrimary: "#111827", TextSecondary: "#6b7280", TextMuted: "#94a3b8", LegendFill: "#6b7280",
		ArrowColors: map[string]string{"control": "#7c3aed", "write": "#10b981", "read": "#2563eb", "data": "#f97316", "async": "#7c3aed", "feedback": "#ef4444", "neutral": "#6b7280"},
	},
	2: {
		Name: "Dark Terminal", FontFamily: "'SF Mono', 'Fira Code', Menlo, monospace",
		Background: "#0f172a", Shadow: false, TitleAlign: "center", TitleFill: "#e2e8f0", TitleSize: 30,
		SubtitleFill: "#94a3b8", SubtitleSize: 14,
		NodeFill: "#111827", NodeStroke: "#334155", NodeRadius: 10, NodeShadow: "",
		SectionFill: "rgba(15,23,42,0.28)", SectionStroke: "#334155", SectionDash: "7 6",
		SectionLabelFill: "#38bdf8", SectionSubFill: "#64748b", SectionUpper: true,
		ArrowWidth: 2.3, ArrowLabelBG: "#0f172a", ArrowLabelOpac: 0.92, ArrowLabelFill: "#cbd5e1",
		TypeLabelFill: "#64748b", TypeLabelSize: 12, TextPrimary: "#e2e8f0", TextSecondary: "#94a3b8", TextMuted: "#64748b", LegendFill: "#94a3b8",
		ArrowColors: map[string]string{"control": "#a855f7", "write": "#22c55e", "read": "#38bdf8", "data": "#fb7185", "async": "#f59e0b", "feedback": "#f97316", "neutral": "#94a3b8"},
	},
	3: {
		Name: "Blueprint", FontFamily: "'SF Mono', 'Fira Code', Menlo, monospace",
		Background: "#082f49", Shadow: false, TitleAlign: "center", TitleFill: "#e0f2fe", TitleSize: 30,
		SubtitleFill: "#7dd3fc", SubtitleSize: 14,
		NodeFill: "#0b3b5e", NodeStroke: "#67e8f9", NodeRadius: 8, NodeShadow: "",
		SectionFill: "none", SectionStroke: "#0ea5e9", SectionDash: "6 4",
		SectionLabelFill: "#67e8f9", SectionSubFill: "#7dd3fc", SectionUpper: true,
		ArrowWidth: 2.1, ArrowLabelBG: "#082f49", ArrowLabelOpac: 0.9, ArrowLabelFill: "#e0f2fe",
		TypeLabelFill: "#7dd3fc", TypeLabelSize: 11, TextPrimary: "#e0f2fe", TextSecondary: "#bae6fd", TextMuted: "#7dd3fc", LegendFill: "#bae6fd",
		ArrowColors: map[string]string{"control": "#67e8f9", "write": "#22d3ee", "read": "#38bdf8", "data": "#fde047", "async": "#c084fc", "feedback": "#fb7185", "neutral": "#bae6fd"},
	},
	4: {
		Name: "Notion Clean", FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', sans-serif",
		Background: "#ffffff", Shadow: false, TitleAlign: "left", TitleFill: "#111827", TitleSize: 18,
		SubtitleFill: "#9ca3af", SubtitleSize: 13,
		NodeFill: "#f9fafb", NodeStroke: "#e5e7eb", NodeRadius: 4, NodeShadow: "",
		SectionFill: "none", SectionStroke: "#e5e7eb", SectionDash: "",
		SectionLabelFill: "#9ca3af", SectionSubFill: "#d1d5db", SectionUpper: true, TitleDivider: true,
		ArrowWidth: 1.8, ArrowLabelBG: "#ffffff", ArrowLabelOpac: 0.96, ArrowLabelFill: "#6b7280",
		TypeLabelFill: "#9ca3af", TypeLabelSize: 11, TextPrimary: "#111827", TextSecondary: "#374151", TextMuted: "#9ca3af", LegendFill: "#6b7280",
		ArrowColors: map[string]string{"control": "#3b82f6", "write": "#3b82f6", "read": "#3b82f6", "data": "#3b82f6", "async": "#9ca3af", "feedback": "#9ca3af", "neutral": "#d1d5db"},
	},
	5: {
		Name: "Glassmorphism", FontFamily: "'Helvetica Neue', Helvetica, Arial, 'PingFang SC', sans-serif",
		Background: "#0f172a", Shadow: true, TitleAlign: "center", TitleFill: "#f8fafc", TitleSize: 30,
		SubtitleFill: "#cbd5e1", SubtitleSize: 14,
		NodeFill: "rgba(255,255,255,0.12)", NodeStroke: "rgba(255,255,255,0.28)", NodeRadius: 18, NodeShadow: "url(#shadowGlass)",
		SectionFill: "rgba(255,255,255,0.05)", SectionStroke: "rgba(255,255,255,0.18)", SectionDash: "7 6",
		SectionLabelFill: "#e2e8f0", SectionSubFill: "#94a3b8", SectionUpper: true,
		ArrowWidth: 2.2, ArrowLabelBG: "#0f172a", ArrowLabelOpac: 0.92, ArrowLabelFill: "#e2e8f0",
		TypeLabelFill: "#94a3b8", TypeLabelSize: 12, TextPrimary: "#f8fafc", TextSecondary: "#cbd5e1", TextMuted: "#94a3b8", LegendFill: "#cbd5e1",
		ArrowColors: map[string]string{"control": "#c084fc", "write": "#34d399", "read": "#60a5fa", "data": "#fb923c", "async": "#f472b6", "feedback": "#f59e0b", "neutral": "#94a3b8"},
	},
	6: {
		Name: "Claude Official", FontFamily: "'Helvetica Neue', Helvetica, Arial, 'PingFang SC', sans-serif",
		Background: "#f8f6f3", Shadow: false, TitleAlign: "center", TitleFill: "#1a1a1a", TitleSize: 26,
		SubtitleFill: "#6b6b6b", SubtitleSize: 14,
		NodeFill: "#ffffff", NodeStroke: "#e0ddd8", NodeRadius: 12, NodeShadow: "",
		SectionFill: "none", SectionStroke: "#e0ddd8", SectionDash: "6 4",
		SectionLabelFill: "#6b6b6b", SectionSubFill: "#a0a0a0", SectionUpper: true,
		ArrowWidth: 2.0, ArrowLabelBG: "#f8f6f3", ArrowLabelOpac: 0.94, ArrowLabelFill: "#6b6b6b",
		TypeLabelFill: "#a0a0a0", TypeLabelSize: 12, TextPrimary: "#1a1a1a", TextSecondary: "#6b6b6b", TextMuted: "#a0a0a0", LegendFill: "#6b6b6b",
		ArrowColors: map[string]string{"control": "#d97757", "write": "#d97757", "read": "#6b6b6b", "data": "#a0a0a0", "async": "#a0a0a0", "feedback": "#d97757", "neutral": "#c0c0c0"},
	},
	7: {
		Name: "OpenAI", FontFamily: "'Helvetica Neue', Helvetica, Arial, 'PingFang SC', sans-serif",
		Background: "#ffffff", Shadow: false, TitleAlign: "left", TitleFill: "#0f172a", TitleSize: 24,
		SubtitleFill: "#64748b", SubtitleSize: 13,
		NodeFill: "#ffffff", NodeStroke: "#dce5e3", NodeRadius: 14, NodeShadow: "",
		SectionFill: "none", SectionStroke: "#e2e8f0", SectionDash: "5 4",
		SectionLabelFill: "#10a37f", SectionSubFill: "#94a3b8", SectionUpper: true, TitleDivider: true,
		ArrowWidth: 2.0, ArrowLabelBG: "#ffffff", ArrowLabelOpac: 0.96, ArrowLabelFill: "#475569",
		TypeLabelFill: "#94a3b8", TypeLabelSize: 11, TextPrimary: "#0f172a", TextSecondary: "#64748b", TextMuted: "#94a3b8", LegendFill: "#64748b",
		ArrowColors: map[string]string{"control": "#10a37f", "write": "#0f766e", "read": "#0891b2", "data": "#f59e0b", "async": "#64748b", "feedback": "#475569", "neutral": "#94a3b8"},
	},
	8: {
		Name: "Dark Luxury", FontFamily: "'Helvetica Neue', Helvetica, Arial, 'PingFang SC', serif",
		Background: "#0a0a0a", Shadow: true, TitleAlign: "center", TitleFill: "#f5e6c8", TitleSize: 30,
		SubtitleFill: "#a0916e", SubtitleSize: 14,
		NodeFill: "#141414", NodeStroke: "#c9a84c", NodeRadius: 8, NodeShadow: "url(#shadowGold)",
		SectionFill: "none", SectionStroke: "#2a2a2a", SectionDash: "5 3",
		SectionLabelFill: "#c9a84c", SectionSubFill: "#6b5d3e", SectionUpper: true,
		ArrowWidth: 2.0, ArrowLabelBG: "#0a0a0a", ArrowLabelOpac: 0.94, ArrowLabelFill: "#c9a84c",
		TypeLabelFill: "#6b5d3e", TypeLabelSize: 12, TextPrimary: "#f5e6c8", TextSecondary: "#a0916e", TextMuted: "#6b5d3e", LegendFill: "#a0916e",
		ArrowColors: map[string]string{"control": "#c9a84c", "write": "#c9a84c", "read": "#a0916e", "data": "#6b5d3e", "async": "#6b5d3e", "feedback": "#c9a84c", "neutral": "#3a3020"},
	},
	9: {
		Name: "C4 Review Canvas", FontFamily: "'Avenir Next', Avenir, 'Segoe UI', 'PingFang SC', sans-serif",
		Background: "#f7f2e8", Shadow: false, TitleAlign: "left", TitleFill: "#24312f", TitleSize: 27, SubtitleFill: "#6f756f", SubtitleSize: 13,
		NodeFill: "#fffdf7", NodeStroke: "#365f56", NodeRadius: 7, NodeShadow: "", SectionFill: "rgba(255,253,247,0.48)", SectionStroke: "#8c7d68", SectionDash: "9 6", SectionLabelFill: "#5b5144", SectionSubFill: "#8c7d68", SectionUpper: false,
		ArrowWidth: 2.0, ArrowLabelBG: "#f7f2e8", ArrowLabelOpac: 0.94, ArrowLabelFill: "#4b5563", TypeLabelFill: "#8a6f43", TypeLabelSize: 10, TextPrimary: "#24312f", TextSecondary: "#5f665f", TextMuted: "#8a8d86", LegendFill: "#5f665f",
		ArrowColors: map[string]string{"control": "#365f56", "write": "#a44a3f", "read": "#356a8a", "data": "#c06b35", "async": "#7a5c99", "feedback": "#b13e53", "neutral": "#746b60"},
	},
	10: {
		Name: "Cloud Fabric", FontFamily: "Inter, 'Helvetica Neue', 'Segoe UI', 'PingFang SC', sans-serif",
		Background: "#edf5fb", Shadow: true, TitleAlign: "left", TitleFill: "#102a43", TitleSize: 27,
		SubtitleFill: "#52718d", SubtitleSize: 13,
		NodeFill: "#ffffff", NodeStroke: "#9bb7cf", NodeRadius: 12, NodeShadow: "url(#shadowSoft)",
		SectionFill: "rgba(255,255,255,0.54)", SectionStroke: "#7fa3c2", SectionDash: "7 5",
		SectionLabelFill: "#315d7e", SectionSubFill: "#7892a8", SectionUpper: true,
		ArrowWidth: 2.2, ArrowLabelBG: "#f7fbfe", ArrowLabelOpac: 0.96, ArrowLabelFill: "#334e68",
		TypeLabelFill: "#6b879d", TypeLabelSize: 10, TextPrimary: "#102a43", TextSecondary: "#486581", TextMuted: "#829ab1", LegendFill: "#486581",
		ArrowColors: map[string]string{"control": "#2563eb", "write": "#ea580c", "read": "#0891b2", "data": "#059669", "async": "#7c3aed", "feedback": "#db2777", "neutral": "#64748b"},
	},
	11: {
		Name: "Event Transit", FontFamily: "'Avenir Next', Avenir, 'Segoe UI', 'PingFang SC', sans-serif",
		Background: "#fbf7ee", Shadow: false, TitleAlign: "left", TitleFill: "#17213c", TitleSize: 27,
		SubtitleFill: "#6e6a61", SubtitleSize: 13, NodeFill: "#fffdf8", NodeStroke: "#c9c2b4", NodeRadius: 24, NodeShadow: "",
		SectionFill: "rgba(255,255,255,0.38)", SectionStroke: "#d4cbbb", SectionDash: "4 5", SectionLabelFill: "#514c43", SectionSubFill: "#8d867b", SectionUpper: true,
		ArrowWidth: 2.8, ArrowLabelBG: "#fbf7ee", ArrowLabelOpac: 0.96, ArrowLabelFill: "#4b5563", TypeLabelFill: "#7a746a", TypeLabelSize: 10, TextPrimary: "#17213c", TextSecondary: "#5e5a52", TextMuted: "#8d867b", LegendFill: "#5e5a52",
		ArrowColors: map[string]string{"control": "#e4475b", "write": "#00897b", "read": "#2563eb", "data": "#f59e0b", "async": "#7c3aed", "feedback": "#c62828", "neutral": "#7a746a"},
	},
	12: {
		Name: "Ops Pulse", FontFamily: "'SF Mono', 'Fira Code', Menlo, monospace",
		Background: "#07111f", Shadow: false, TitleAlign: "left", TitleFill: "#eff6ff", TitleSize: 27, SubtitleFill: "#8aa4bd", SubtitleSize: 13,
		NodeFill: "#0d1b2a", NodeStroke: "#29435d", NodeRadius: 12, NodeShadow: "", SectionFill: "rgba(13,27,42,0.72)", SectionStroke: "#28445f", SectionDash: "6 5", SectionLabelFill: "#38bdf8", SectionSubFill: "#6f8ba5", SectionUpper: true,
		ArrowWidth: 2.4, ArrowLabelBG: "#07111f", ArrowLabelOpac: 0.94, ArrowLabelFill: "#cbd5e1", TypeLabelFill: "#6f8ba5", TypeLabelSize: 10, TextPrimary: "#eff6ff", TextSecondary: "#9fb3c8", TextMuted: "#647f99", LegendFill: "#9fb3c8",
		ArrowColors: map[string]string{"control": "#f59e0b", "write": "#22c55e", "read": "#38bdf8", "data": "#fb7185", "async": "#22d3ee", "feedback": "#f43f5e", "neutral": "#7892a8"},
	},
}

func GetProfile(styleIndex int) *StyleProfile {
	if p, ok := StyleProfiles[styleIndex]; ok {
		return p
	}
	return StyleProfiles[1] // default Flat Icon
}

// --- SVG Builder ---

type builder struct {
	lines []string
}

func (b *builder) add(format string, args ...any) {
	b.lines = append(b.lines, fmt.Sprintf(format, args...))
}

func (b *builder) join() string {
	return strings.Join(b.lines, "\n")
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

var reservedDOMIDs = map[string]struct{}{
	"blueprint-title-block": {}, "blueprintGrid": {}, "cloudGradient": {}, "cloudGrid": {},
	"footer": {}, "glowBlue": {}, "glowGreen": {}, "glowOrange": {}, "glowPurple": {},
	"legend": {}, "opsGradient": {}, "opsGrid": {}, "pulseGlow": {}, "reviewGrid": {},
	"shadowGlass": {}, "shadowGold": {}, "shadowSoft": {}, "style-signature": {},
	"terminalGradient": {}, "transitDots": {},
}

func safeDOMID(value, fallback string) string {
	var out strings.Builder
	previousDash := false
	for _, r := range value {
		allowed := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '.' || r == ':' || r == '-'
		if allowed {
			out.WriteRune(r)
			previousDash = r == '-'
			continue
		}
		if out.Len() > 0 && !previousDash {
			out.WriteByte('-')
			previousDash = true
		}
	}
	if cleaned := strings.Trim(out.String(), "-"); cleaned != "" {
		return cleaned
	}
	return fallback
}

func allocateDOMID(base, fallback string, used map[string]struct{}, suffixes ...string) string {
	base = safeDOMID(base, fallback)
	if base == "" {
		base = fallback
	}
	for sequence := 1; ; sequence++ {
		candidate := base
		if sequence > 1 {
			candidate = fmt.Sprintf("%s-%d", base, sequence)
		}
		if _, collision := used[candidate]; collision {
			continue
		}
		collision := false
		for _, suffix := range suffixes {
			if _, exists := used[candidate+suffix]; exists {
				collision = true
				break
			}
		}
		if collision {
			continue
		}
		used[candidate] = struct{}{}
		for _, suffix := range suffixes {
			used[candidate+suffix] = struct{}{}
		}
		return candidate
	}
}

func allocateDiagramDOMIDs(d *ir.Diagram, profile *StyleProfile) (map[string]string, map[string]string, map[string]string) {
	used := make(map[string]struct{}, len(reservedDOMIDs)+len(profile.ArrowColors)+len(d.Containers)+len(d.Nodes)+len(d.Edges))
	for id := range reservedDOMIDs {
		used[id] = struct{}{}
	}
	for flow := range profile.ArrowColors {
		used["arrow-"+flow] = struct{}{}
	}
	containerIDs := make(map[string]string, len(d.Containers))
	for _, c := range d.Containers {
		containerIDs[c.ID] = allocateDOMID("container-"+safeDOMID(c.ID, "container"), "container", used, "-header")
	}
	edgeIDs := make(map[string]string, len(d.Edges))
	for _, e := range d.Edges {
		edgeIDs[e.ID] = allocateDOMID(safeDOMID(e.ID, "edge"), "edge", used, "-critical-glow", "-hop", "-label")
	}
	nodeIDs := make(map[string]string, len(d.Nodes))
	for _, n := range d.Nodes {
		nodeIDs[n.ID] = allocateDOMID("node-"+safeDOMID(n.ID, "node"), "node", used)
	}
	return containerIDs, edgeIDs, nodeIDs
}

func semanticRole(raw map[string]any, fallback string) string {
	for _, key := range []string{"c4_type", "deployment_kind", "transit_role", "ops_role", "kind"} {
		if value := rawMapString(raw, key); value != "" {
			return value
		}
	}
	return fallback
}

// --- Main Entry Point ---

// RenderSVG generates an SVG string from a normalized Diagram.
func RenderSVG(d *ir.Diagram, profile *StyleProfile) string {
	b := &builder{}

	width := d.Width
	height := d.Height
	if width <= 0 {
		width = 960
	}
	if height <= 0 {
		height = 600
	}

	styleIndex := d.StyleIndex
	containerDOMIDs, edgeDOMIDs, nodeDOMIDs := allocateDiagramDOMIDs(d, profile)

	b.add(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f" data-generator="drawcli-go" data-schema-version="1" data-text-metrics="heuristic-v1" data-style-id="%d" data-visual-theme="%s" data-diagram-type="%s" data-motion-scene="%s" data-semantic-profile="%s" data-semantic-valid="true"%s>`,
		width, height, width, height, styleIndex, esc(profile.Name), esc(d.Mode), esc(d.MotionScene), esc(d.SemanticProfile), compositionContractAttributes(d.QualityProfile))

	// --- defs ---
	b.add(`<defs>`)
	// Filters
	if profile.Shadow {
		b.add(`  <filter id="shadowSoft" x="-20%%" y="-20%%" width="140%%" height="160%%">`)
		b.add(`    <feDropShadow dx="0" dy="3" stdDeviation="6" flood-color="#0f172a" flood-opacity="0.12"/>`)
		b.add(`  </filter>`)
		b.add(`  <filter id="shadowGlass" x="-20%%" y="-20%%" width="140%%" height="160%%">`)
		b.add(`    <feDropShadow dx="0" dy="10" stdDeviation="16" flood-color="#020617" flood-opacity="0.28"/>`)
		b.add(`  </filter>`)
	}
	if styleIndex == 8 {
		b.add(`  <filter id="shadowGold" x="-20%%" y="-20%%" width="140%%" height="160%%">`)
		b.add(`    <feDropShadow dx="0" dy="3" stdDeviation="6" flood-color="#c9a84c" flood-opacity="0.18"/>`)
		b.add(`  </filter>`)
	}
	// Style-specific gradients/patterns
	if styleIndex == 2 {
		b.add(`  <linearGradient id="terminalGradient" x1="0%%" y1="0%%" x2="100%%" y2="100%%">`)
		b.add(`    <stop offset="0%%" stop-color="#0f0f1a"/>`)
		b.add(`    <stop offset="100%%" stop-color="#1a1a2e"/>`)
		b.add(`  </linearGradient>`)
		for _, g := range []struct{ id, color string }{{"glowBlue", "#3b82f6"}, {"glowPurple", "#a855f7"}, {"glowGreen", "#22c55e"}, {"glowOrange", "#f97316"}} {
			b.add(`  <filter id="%s" x="-30%%" y="-30%%" width="160%%" height="160%%">`, g.id)
			b.add(`    <feDropShadow dx="0" dy="0" stdDeviation="5" flood-color="%s" flood-opacity="0.65"/>`, g.color)
			b.add(`  </filter>`)
		}
	}
	if styleIndex == 3 {
		b.add(`  <pattern id="blueprintGrid" width="32" height="32" patternUnits="userSpaceOnUse">`)
		b.add(`    <path d="M 32 0 L 0 0 0 32" fill="none" stroke="#0ea5e9" stroke-opacity="0.12" stroke-width="1"/>`)
		b.add(`  </pattern>`)
	}
	if styleIndex == 9 {
		b.add(`  <pattern id="reviewGrid" width="24" height="24" patternUnits="userSpaceOnUse">`)
		b.add(`    <circle cx="1" cy="1" r="0.8" fill="#8c7d68" fill-opacity="0.18"/>`)
		b.add(`  </pattern>`)
	}
	if styleIndex == 10 {
		b.add(`  <linearGradient id="cloudGradient" x1="0%%" y1="0%%" x2="100%%" y2="100%%">`)
		b.add(`    <stop offset="0%%" stop-color="#f8fcff"/>`)
		b.add(`    <stop offset="100%%" stop-color="#dfedf7"/>`)
		b.add(`  </linearGradient>`)
		b.add(`  <pattern id="cloudGrid" width="32" height="32" patternUnits="userSpaceOnUse">`)
		b.add(`    <path d="M 32 0 L 0 0 0 32" fill="none" stroke="#7fa3c2" stroke-opacity="0.10" stroke-width="1"/>`)
		b.add(`  </pattern>`)
	}
	if styleIndex == 11 {
		b.add(`  <pattern id="transitDots" width="28" height="28" patternUnits="userSpaceOnUse">`)
		b.add(`    <circle cx="2" cy="2" r="0.9" fill="#8d867b" fill-opacity="0.12"/>`)
		b.add(`  </pattern>`)
	}
	if styleIndex == 12 {
		b.add(`  <linearGradient id="opsGradient" x1="0%%" y1="0%%" x2="100%%" y2="100%%">`)
		b.add(`    <stop offset="0%%" stop-color="#07111f"/>`)
		b.add(`    <stop offset="100%%" stop-color="#0b1b2e"/>`)
		b.add(`  </linearGradient>`)
		b.add(`  <pattern id="opsGrid" width="36" height="36" patternUnits="userSpaceOnUse">`)
		b.add(`    <path d="M 36 0 L 0 0 0 36" fill="none" stroke="#38bdf8" stroke-opacity="0.055" stroke-width="1"/>`)
		b.add(`  </pattern>`)
		b.add(`  <filter id="pulseGlow" x="-30%%" y="-30%%" width="160%%" height="160%%">`)
		b.add(`    <feDropShadow dx="0" dy="0" stdDeviation="4" flood-color="#f59e0b" flood-opacity="0.62"/>`)
		b.add(`  </filter>`)
	}
	// Arrow markers
	for flow, color := range profile.ArrowColors {
		markerID := fmt.Sprintf("arrow-%s", flow)
		b.add(`  <marker id="%s" viewBox="0 0 10 10" refX="9" refY="3.5" markerWidth="10" markerHeight="7" orient="auto">`, markerID)
		b.add(`    <path d="M 0 1 L 9 3.5 L 0 6 Z" fill="%s"/>`, color)
		b.add(`  </marker>`)
	}
	// CSS style classes
	b.add(`  <style>`)
	fontFamily := profile.FontFamily
	b.add(`    text { font-family: %s; }`, fontFamily)
	b.add(`    .title { font-size: %.0fpx; font-weight: 700; fill: %s; }`, profile.TitleSize, profile.TitleFill)
	b.add(`    .subtitle { font-size: %.0fpx; font-weight: 500; fill: %s; }`, profile.SubtitleSize, profile.SubtitleFill)
	b.add(`    .section { font-size: 13px; font-weight: 700; fill: %s; letter-spacing: 1.4px; }`, profile.SectionLabelFill)
	b.add(`    .section-sub { font-size: 12px; font-weight: 500; fill: %s; }`, profile.SectionSubFill)
	b.add(`    .node-title { font-weight: 700; fill: %s; font-size: 13px; }`, profile.TextPrimary)
	b.add(`    .node-type { font-size: %.0fpx; font-weight: 500; fill: %s; letter-spacing: 0.6px; }`, profile.TypeLabelSize, profile.TypeLabelFill)
	b.add(`    .node-sub { font-size: 12px; font-weight: 500; fill: %s; }`, profile.TextSecondary)
	b.add(`    .arrow-label { font-size: 12px; font-weight: 600; fill: %s; }`, profile.ArrowLabelFill)
	b.add(`    .legend { font-size: 12px; font-weight: 500; fill: %s; }`, profile.LegendFill)
	b.add(`    .footnote { font-size: 12px; font-weight: 500; fill: %s; }`, profile.TextMuted)
	b.add(`  </style>`)
	b.add(`</defs>`)

	// --- Canvas background ---
	if styleIndex == 2 {
		b.add(`  <rect data-graph-role="background" width="%.0f" height="%.0f" fill="url(#terminalGradient)"/>`, width, height)
	} else if styleIndex == 10 {
		b.add(`  <rect data-graph-role="background" width="%.0f" height="%.0f" fill="url(#cloudGradient)"/>`, width, height)
		b.add(`  <rect data-graph-role="decoration" width="%.0f" height="%.0f" fill="url(#cloudGrid)"/>`, width, height)
	} else if styleIndex == 12 {
		b.add(`  <rect data-graph-role="background" width="%.0f" height="%.0f" fill="url(#opsGradient)"/>`, width, height)
		b.add(`  <rect data-graph-role="decoration" width="%.0f" height="%.0f" fill="url(#opsGrid)"/>`, width, height)
	} else {
		b.add(`  <rect data-graph-role="background" width="%.0f" height="%.0f" fill="%s"/>`, width, height, profile.Background)
		if styleIndex == 3 {
			b.add(`  <rect data-graph-role="decoration" width="%.0f" height="%.0f" fill="url(#blueprintGrid)"/>`, width, height)
		} else if styleIndex == 9 {
			b.add(`  <rect data-graph-role="decoration" width="%.0f" height="%.0f" fill="url(#reviewGrid)"/>`, width, height)
		} else if styleIndex == 11 {
			b.add(`  <rect data-graph-role="decoration" width="%.0f" height="%.0f" fill="url(#transitDots)"/>`, width, height)
		}
	}

	// --- Window controls (Style 2) ---
	if styleIndex == 2 {
		for i, c := range []string{"#ef4444", "#f59e0b", "#10b981"} {
			b.add(`  <circle cx="%.0f" cy="20" r="5.5" fill="%s"/>`, 20+float64(i)*18, c)
		}
	}

	// --- Header meta ---
	if d.MetaLeft != "" || d.MetaCenter != "" || d.MetaRight != "" {
		renderHeaderMeta(b, d, profile, width)
	}

	// --- Style signature badge (Style 9-12) ---
	renderStyleSignature(b, d, profile, width, height)

	// --- Containers ---
	for _, c := range d.Containers {
		renderContainer(b, c, containerDOMIDs[c.ID], profile, styleIndex)
	}

	// --- Arrows (before nodes for z-order) ---
	arrowRoutes := buildArrowRoutes(d, profile)
	for index := range arrowRoutes {
		arrowRoutes[index].DOMID = edgeDOMIDs[arrowRoutes[index].Edge.ID]
	}
	for _, ar := range arrowRoutes {
		renderArrow(b, ar, profile)
	}

	// --- Nodes ---
	for _, n := range d.Nodes {
		b.add(`<g id="%s" data-graph-role="node" data-node-id="%s" data-semantic-role="%s" data-motion-role="%s" data-motion-stage="%s" data-motion-order="%s" data-parent="%s" data-deployment-id="%s" data-topic-id="%s" data-span-id="%s" data-station-order="%s" data-status="%s" data-start-ms="%s" data-duration-ms="%s" data-parent-span="%s" data-graph-bounds="%.0f,%.0f,%.0f,%.0f">`,
			esc(nodeDOMIDs[n.ID]), esc(n.ID), esc(semanticRole(n.Raw, n.Kind)), esc(rawMapString(n.Raw, "motion_role")), esc(rawMapString(n.Raw, "motion_stage")), esc(rawMapString(n.Raw, "motion_order")), esc(rawMapString(n.Raw, "parent")), esc(rawMapString(n.Raw, "deployment_id")), esc(rawMapString(n.Raw, "topic_id")), esc(rawMapString(n.Raw, "span_id")), esc(rawMapString(n.Raw, "station_order")), esc(rawMapString(n.Raw, "status")), esc(rawMapString(n.Raw, "start_ms")), esc(rawMapString(n.Raw, "duration_ms")), esc(rawMapString(n.Raw, "parent_span")), n.X, n.Y, n.X+n.Width, n.Y+n.Height)
		renderNode(b, n, profile)
		b.add(`</g>`)
	}

	// --- Title ---
	renderTitleBlock(b, d, profile, width)

	// --- Blueprint title block (Style 3) ---
	if styleIndex == 3 && d.Title != "" {
		renderBlueprintTitleBlock(b, d, profile, width, height)
	}

	// --- Legend ---
	renderLegend(b, d, profile)

	// --- Footer ---
	if d.Footer != "" {
		fy := d.FooterY
		if fy == 0 {
			fy = height - 20
		}
		fx := d.FooterX
		if fx == 0 {
			fx = 48
		}
		b.add(`<text x="%.0f" y="%.0f" class="footnote">%s</text>`, fx, fy, esc(d.Footer))
	}

	b.add(`</svg>`)
	return b.join()
}

func compositionContractAttributes(profile string) string {
	if strings.EqualFold(strings.TrimSpace(profile), "showcase") {
		return ` data-quality-profile="showcase" data-max-bends-per-edge="2" data-max-total-bends="8" data-max-route-stretch="1.35" data-max-bridged-crossings="0" data-min-node-gap="40" data-min-container-gutter="20" data-min-label-clearance="4" data-min-segment-length="16"`
	}
	return ` data-quality-profile="standard" data-max-bends-per-edge="12" data-max-total-bends="100" data-max-route-stretch="5" data-max-bridged-crossings="8" data-min-node-gap="0" data-min-container-gutter="0" data-min-label-clearance="2" data-min-segment-length="0"`
}

// --- Header Meta ---

func renderHeaderMeta(b *builder, d *ir.Diagram, profile *StyleProfile, width float64) {
	fill := profile.TextMuted
	var size float64 = 11
	if d.MetaLeft != "" {
		b.add(`  <text x="28" y="24" font-size="%.0f" font-weight="600" fill="%s">%s</text>`, size, fill, esc(d.MetaLeft))
	}
	if d.MetaCenter != "" {
		b.add(`  <text x="%.0f" y="24" text-anchor="middle" font-size="%.0f" font-weight="600" fill="%s">%s</text>`, width/2, size, fill, esc(d.MetaCenter))
	}
	if d.MetaRight != "" {
		b.add(`  <text x="%.0f" y="24" text-anchor="end" font-size="%.0f" font-weight="600" fill="%s">%s</text>`, width-28, size, fill, esc(d.MetaRight))
	}
}

// --- Style Signature (Style 9-12) ---

func renderStyleSignature(b *builder, d *ir.Diagram, profile *StyleProfile, width, _ float64) {
	styleIndex := d.StyleIndex
	if styleIndex < 9 || styleIndex > 12 {
		return
	}
	badgeW := 176.0
	badgeH := 34.0
	x := width - 48 - badgeW
	y := 22.0

	switch styleIndex {
	case 9:
		level := d.C4Level
		if level == "" {
			level = "REVIEW"
		}
		state := d.ReviewState
		if state == "" {
			state = "REVIEW READY"
		}
		b.add(`  <g id="style-signature" data-graph-role="decoration" data-style-signature="c4-review-board">`)
		b.add(`    <rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="7" fill="#fffdf7" stroke="#8c7d68" stroke-width="1.2" stroke-dasharray="6 4"/>`, x, y, badgeW, badgeH)
		b.add(`    <path d="M %.0f %.0f l 5 5 9 -11" fill="none" stroke="#365f56" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`, x+12, y+17)
		b.add(`    <text x="%.0f" y="%.0f" font-size="8.5" font-weight="800" fill="#8a6f43">C4 · %s VIEW</text>`, x+34, y+14, esc(level))
		b.add(`    <text x="%.0f" y="%.0f" font-size="8" font-weight="700" fill="#5f665f">%s</text>`, x+34, y+27, esc(state))
		b.add(`  </g>`)
	case 10:
		platform := d.PlatformProfile
		if platform == "" {
			platform = "CLOUD"
		}
		regionCount := 0
		for _, c := range d.Containers {
			if c.DeploymentKind == "region" {
				regionCount++
			}
		}
		b.add(`  <g id="style-signature" data-graph-role="decoration" data-style-signature="cloud-ownership-map">`)
		b.add(`    <rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="9" fill="#ffffff" fill-opacity="0.82" stroke="#7fa3c2" stroke-width="1.1"/>`, x, y, badgeW, badgeH)
		b.add(`    <rect x="%.0f" y="%.0f" width="14" height="14" rx="4" fill="#dbeafe" stroke="#2563eb" stroke-width="1"/>`, x+11, y+9)
		b.add(`    <rect x="%.0f" y="%.0f" width="14" height="14" rx="4" fill="#dcfce7" stroke="#059669" stroke-width="1"/>`, x+20, y+13)
		b.add(`    <text x="%.0f" y="%.0f" font-size="8.5" font-weight="800" fill="#315d7e">%s · %d REGIONS</text>`, x+43, y+14, esc(platform), regionCount)
		b.add(`    <text x="%.0f" y="%.0f" font-size="8" font-weight="700" fill="#52718d">DEPLOYMENT MAP</text>`, x+43, y+27)
		b.add(`  </g>`)
	case 11:
		topicCount := len(d.Topics)
		b.add(`  <g id="style-signature" data-graph-role="decoration" data-style-signature="event-metro-map">`)
		b.add(`    <rect x="%.0f" y="%.0f" width="226" height="%.0f" rx="7" fill="#17213c" stroke="#514c43" stroke-width="1"/>`, x-50, y, badgeH)
		b.add(`    <line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#e4475b" stroke-width="3"/>`, x+12-50, y+17, x+36-50, y+17)
		b.add(`    <circle cx="%.0f" cy="%.0f" r="4" fill="#fbf7ee" stroke="#e4475b" stroke-width="2"/>`, x+18-50, y+17)
		b.add(`    <circle cx="%.0f" cy="%.0f" r="4" fill="#fbf7ee" stroke="#e4475b" stroke-width="2"/>`, x+31-50, y+17)
		b.add(`    <text x="%.0f" y="%.0f" font-size="8.5" font-weight="800" fill="#ffffff">EVENT METRO</text>`, x-4, y+14)
		b.add(`    <text x="%.0f" y="%.0f" font-size="8" font-weight="700" fill="#f3d5d9">%d TOPIC LINES · DECLARED STOPS</text>`, x-4, y+27, topicCount)
		b.add(`  </g>`)
	case 12:
		window := d.ObservationWindow
		if window == "" {
			window = "WINDOW"
		}
		worst := "unknown"
		worstRank := 0
		for _, node := range d.Nodes {
			if rawString(node, "ops_role") != "service" {
				continue
			}
			status := strings.ToLower(rawString(node, "status"))
			rank := map[string]int{"ok": 1, "warn": 2, "critical": 3}[status]
			if rank > worstRank {
				worst = status
				worstRank = rank
			}
		}
		statusColor := opsStatusColor(worst)
		b.add(`  <g id="style-signature" data-graph-role="decoration" data-style-signature="ops-live-investigation">`)
		b.add(`    <rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="7" fill="#0d1b2a" stroke="#29435d" stroke-width="1.1"/>`, x, y, badgeW, badgeH)
		b.add(`    <circle cx="%.0f" cy="%.0f" r="4" fill="%s"/>`, x+15, y+17, statusColor)
		b.add(`    <path d="M %.0f %.0f h 5 l 3 -6 5 12 4 -8 h 7" fill="none" stroke="#38bdf8" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>`, x+25, y+18)
		b.add(`    <text x="%.0f" y="%.0f" font-size="8.5" font-weight="800" fill="#eff6ff">LIVE · %s</text>`, x+58, y+14, esc(window))
		b.add(`    <text x="%.0f" y="%.0f" font-size="8" font-weight="700" fill="%s">%s · CORRELATED TRACE</text>`, x+58, y+27, statusColor, esc(strings.ToUpper(worst)))
		b.add(`  </g>`)
	}
}

// --- Blueprint Title Block (Style 3) ---

func renderBlueprintTitleBlock(b *builder, d *ir.Diagram, profile *StyleProfile, width, height float64) {
	bw := 256.0
	bh := 92.0
	bx := width - bw - 28
	by := height - bh - 18
	block, _ := d.Raw["blueprint_title_block"].(map[string]any)
	if block == nil {
		return
	}
	bw = rawMapFloat(block, "width", bw)
	bh = rawMapFloat(block, "height", bh)
	bx = rawMapFloat(block, "x", width-bw-28)
	by = rawMapFloat(block, "y", height-bh-18)
	title := esc(rawMapString(block, "title"))
	if title == "" {
		title = esc(d.Title)
	}
	subtitle := rawMapString(block, "subtitle")
	if subtitle == "" {
		subtitle = "SYSTEM ARCHITECTURE"
	}
	leftCaption := rawMapString(block, "left_caption")
	if leftCaption == "" {
		leftCaption = "REV: 1.0"
	}
	centerCaption := rawMapString(block, "center_caption")
	if centerCaption == "" {
		centerCaption = "AUTO-GENERATED"
	}
	rightCaption := rawMapString(block, "right_caption")
	if rightCaption == "" {
		rightCaption = "DWG: ARCH-001"
	}
	stroke := profile.SectionStroke
	fill := "#0b3552"

	b.add(`  <g id="blueprint-title-block" data-graph-role="decoration">`)
	b.add(`    <rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="%s" stroke="%s" stroke-width="1.2"/>`, bx, by, bw, bh, fill, stroke)
	b.add(`    <line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="1"/>`, bx, by+18, bx+bw, by+18, stroke)
	b.add(`    <line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="1"/>`, bx, by+54, bx+bw, by+54, stroke)
	colW := bw / 3
	b.add(`    <line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="0.7"/>`, bx+colW, by+54, bx+colW, by+bh, stroke)
	b.add(`    <line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="0.7"/>`, bx+colW*2, by+54, bx+colW*2, by+bh, stroke)
	b.add(`    <text x="%.0f" y="%.0f" text-anchor="middle" font-size="10" font-weight="600" fill="%s">%s</text>`, bx+bw/2, by+13, profile.TextMuted, esc(subtitle))
	b.add(`    <text x="%.0f" y="%.0f" text-anchor="middle" font-size="18" font-weight="700" fill="%s">%s</text>`, bx+bw/2, by+42, profile.TextPrimary, title)
	captionY := by + bh - 4
	b.add(`    <text x="%.0f" y="%.0f" text-anchor="middle" font-size="9.5" font-weight="600" fill="%s">%s</text>`, bx+colW/2, captionY, profile.TextMuted, esc(leftCaption))
	b.add(`    <text x="%.0f" y="%.0f" text-anchor="middle" font-size="9.5" font-weight="600" fill="%s">%s</text>`, bx+bw/2, captionY, profile.SectionLabelFill, esc(centerCaption))
	b.add(`    <text x="%.0f" y="%.0f" text-anchor="middle" font-size="9.5" font-weight="600" fill="%s">%s</text>`, bx+bw-colW/2, captionY, profile.TextMuted, esc(rightCaption))
	b.add(`  </g>`)
}

// --- Container Rendering ---

func renderContainer(b *builder, c ir.Container, domID string, profile *StyleProfile, styleIndex int) {
	rx := 16.0
	if styleIndex == 4 {
		rx = 4
	}
	dash := ""
	if profile.SectionDash != "" {
		dash = fmt.Sprintf(` stroke-dasharray="%s"`, profile.SectionDash)
	}

	b.add(`  <g id="%s" data-graph-role="container" data-container-id="%s" data-semantic-role="%s" data-graph-bounds="%.0f,%.0f,%.0f,%.0f">`, esc(domID), esc(c.ID), esc(semanticRole(c.Raw, "boundary")), c.X, c.Y, c.X+c.Width, c.Y+c.Height)

	b.add(`  <rect data-graph-role="container-bg" x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="1.4"%s/>`,
		c.X, c.Y, c.Width, c.Height, rx, profile.SectionFill, profile.SectionStroke, dash)

	// Style-specific container decorations
	if styleIndex == 12 {
		b.add(`  <line data-graph-role="decoration" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#38bdf8" stroke-width="1" opacity="0.16"/>`,
			c.X+14, c.Y+34, c.X+c.Width-14, c.Y+34)
		if strings.Contains(strings.ToLower(c.ID), "trace") {
			rulerLeft := c.X + c.Width - 244
			rulerRight := c.X + c.Width - 24
			b.add(`  <line data-graph-role="decoration" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#38bdf8" stroke-width="1" opacity="0.42"/>`, rulerLeft, c.Y+22, rulerRight, c.Y+22)
			for i := range 5 {
				tickX := rulerLeft + (rulerRight-rulerLeft)*float64(i)/4
				b.add(`  <line data-graph-role="decoration" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="#38bdf8" stroke-width="1" opacity="0.52"/>`, tickX, c.Y+18, tickX, c.Y+26)
				b.add(`  <text data-graph-role="decoration" x="%.0f" y="%.0f" text-anchor="middle" font-size="7" font-weight="700" fill="#6f8ba5">%d%%</text>`, tickX, c.Y+14, i*25)
			}
		}
	}
	if styleIndex == 11 {
		b.add(`  <line data-graph-role="decoration" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="1" stroke-dasharray="2 7" opacity="0.35"/>`,
			c.X+18, c.Y+34, c.X+c.Width-18, c.Y+34, profile.SectionStroke)
	}

	// Container label with reserved header zone indicator
	sectionLabel := c.Label
	if prefix := rawMapString(c.Raw, "header_prefix"); prefix != "" {
		separator := rawMapString(c.Raw, "header_separator")
		if separator == "" {
			separator = " // "
		}
		sectionLabel = prefix + separator + sectionLabel
	}
	preserveCase, _ := c.Raw["preserve_case"].(bool)
	if profile.SectionUpper && !preserveCase {
		sectionLabel = strings.ToUpper(c.Label)
		if prefix := rawMapString(c.Raw, "header_prefix"); prefix != "" {
			separator := rawMapString(c.Raw, "header_separator")
			if separator == "" {
				separator = " // "
			}
			sectionLabel = strings.ToUpper(prefix + separator + c.Label)
		}
	}
	hasSub := c.Subtitle != ""
	headerH := 30.0
	if hasSub {
		headerH = 54.0
	}
	if sectionLabel != "" || hasSub {
		// Narrow reserved header to text width, not full container
		labelW := geometry.EstimateTextWidth(sectionLabel)
		if hasSub {
			subW := geometry.EstimateTextWidth(c.Subtitle)
			if subW > labelW {
				labelW = subW
			}
		}
		headerW := math.Min(c.Width-16, labelW+30)
		b.add(`  <rect id="%s-header" data-graph-role="reserved" data-reserved-kind="container-header" x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="none" stroke="none"/>`, esc(domID),
			c.X+8, c.Y+6, headerW, headerH)
	}
	if sectionLabel != "" {
		b.add(`  <text x="%.0f" y="%.0f" class="section">%s</text>`, c.X+18, c.Y+24, esc(sectionLabel))
	}
	if hasSub {
		b.add(`  <text x="%.0f" y="%.0f" class="section-sub">%s</text>`, c.X+18, c.Y+44, esc(c.Subtitle))
	}

	b.add(`  </g>`)
}

// --- Title Block ---

func renderTitleBlock(b *builder, d *ir.Diagram, profile *StyleProfile, width float64) {
	if d.Title == "" {
		return
	}
	var titleX float64
	if profile.TitleAlign == "left" {
		titleX = 48
	} else {
		titleX = width / 2
	}
	anchor := "middle"
	if profile.TitleAlign == "left" {
		anchor = "start"
	}
	titleY := 56.0

	b.add(`<text x="%.0f" y="%.0f" class="title" text-anchor="%s">%s</text>`,
		titleX, titleY, anchor, esc(d.Title))
	if d.Subtitle != "" {
		subY := titleY + 26
		b.add(`<text x="%.0f" y="%.0f" class="subtitle" text-anchor="%s">%s</text>`,
			titleX, subY, anchor, esc(d.Subtitle))
	}
}

// --- Node Rendering ---

func renderNode(b *builder, n ir.Node, profile *StyleProfile) {
	fill := n.Fill
	if fill == "" {
		fill = profile.NodeFill
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = profile.NodeStroke
	}
	radius := profile.NodeRadius
	shadow := profile.NodeShadow

	cx := n.X + n.Width/2
	cy := n.Y + n.Height/2

	shadowStr := ""
	if shadow != "" && !n.Flat {
		shadowStr = fmt.Sprintf(` filter="%s"`, shadow)
	}

	switch n.Kind {
	case "transit_terminal", "transit_station", "transit_junction":
		renderTransitNode(b, n, profile)
	case "cloud_service":
		renderCloudService(b, n, profile)
	case "ops_service":
		renderOpsService(b, n, profile)
	case "otel_collector":
		renderOTelCollector(b, n, profile)
	case "trace_span":
		renderTraceSpan(b, n, profile)
	case "cylinder":
		ellipseRX := n.Width / 4
		ellipseRY := math.Min(18, n.Height/8)
		strokeWidth := 2.2
		b.add(`<ellipse cx="%.0f" cy="%.0f" rx="%.0f" ry="%.0f" fill="%s" stroke="%s" stroke-width="%.1f"%s/>`,
			cx, n.Y+ellipseRY, ellipseRX, ellipseRY, fill, stroke, strokeWidth, shadowStr)
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="%s" stroke="%s" stroke-width="%.1f"/>`,
			n.X, n.Y+ellipseRY, n.Width, n.Height-2*ellipseRY, fill, stroke, strokeWidth)
		b.add(`<ellipse cx="%.0f" cy="%.0f" rx="%.0f" ry="%.0f" fill="%s" stroke="%s" stroke-width="%.1f"/>`,
			cx, n.Y+n.Height-ellipseRY, ellipseRX, ellipseRY, fill, stroke, strokeWidth)
		b.add(`<ellipse cx="%.0f" cy="%.0f" rx="%.0f" ry="%.0f" fill="none" stroke="%s" stroke-opacity="0.45" stroke-width="1.2"/>`,
			cx, n.Y+n.Height*0.38, ellipseRX, ellipseRY, stroke)
		b.add(`<ellipse cx="%.0f" cy="%.0f" rx="%.0f" ry="%.0f" fill="none" stroke="%s" stroke-opacity="0.25" stroke-width="1.2"/>`,
			cx, n.Y+n.Height*0.60, ellipseRX, ellipseRY, stroke)

	case "double_rect":
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="1.8"%s/>`,
			n.X, n.Y, n.Width, n.Height, radius, fill, stroke, shadowStr)
		innerR := radius - 2
		if innerR < 4 {
			innerR = 4
		}
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="none" stroke="%s" stroke-width="1.2" opacity="0.65"/>`,
			n.X+5, n.Y+5, n.Width-10, n.Height-10, innerR, stroke)

	case "hexagon":
		inset := 22.0
		if n.Width < 120 {
			inset = n.Width * 0.15
		}
		pts := fmt.Sprintf("%.0f,%.0f %.0f,%.0f %.0f,%.0f %.0f,%.0f %.0f,%.0f %.0f,%.0f",
			n.X+inset, n.Y, n.X+n.Width-inset, n.Y, n.X+n.Width, cy,
			n.X+n.Width-inset, n.Y+n.Height, n.X+inset, n.Y+n.Height, n.X, cy)
		b.add(`<polygon points="%s" fill="%s" stroke="%s" stroke-width="1.5"%s/>`, pts, fill, stroke, shadowStr)

	case "speech":
		tail := 14.0
		path := fmt.Sprintf("M %.0f %.0f L %.0f %.0f Q %.0f %.0f %.0f %.0f L %.0f %.0f Q %.0f %.0f %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f Q %.0f %.0f %.0f %.0f L %.0f %.0f Q %.0f %.0f %.0f %.0f Z",
			n.X+radius, n.Y,
			n.X+n.Width-radius, n.Y,
			n.X+n.Width, n.Y, n.X+n.Width, n.Y+radius,
			n.X+n.Width, n.Y+n.Height-radius,
			n.X+n.Width, n.Y+n.Height, n.X+n.Width-radius, n.Y+n.Height,
			n.X+26, n.Y+n.Height,
			n.X+12, n.Y+n.Height+tail,
			n.X+16, n.Y+n.Height,
			n.X+radius, n.Y+n.Height,
			n.X, n.Y+n.Height, n.X, n.Y+n.Height-radius,
			n.X, n.Y+radius,
			n.X, n.Y, n.X+radius, n.Y)
		b.add(`<path d="%s" fill="%s" stroke="%s" stroke-width="1.5"%s/>`, path, fill, stroke, shadowStr)

	case "user_avatar":
		outerStrokeWidth := 1.5
		if n.Flat {
			outerStrokeWidth = 2.0
		}
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="%.1f"%s/>`,
			n.X, n.Y, n.Width, n.Height, radius, fill, stroke, outerStrokeWidth, shadowStr)
		circleCX := n.X + 26
		circleCY := n.Y + n.Height/2
		iconStroke := stroke
		iconFill := "#dbeafe"
		if rawIconFill, ok := n.Raw["icon_fill"].(string); ok && rawIconFill != "" {
			iconFill = rawIconFill
		}
		b.add(`<circle cx="%.0f" cy="%.0f" r="18" fill="%s" stroke="%s" stroke-width="1.6"/>`,
			circleCX, circleCY, iconFill, iconStroke)
		b.add(`<circle cx="%.0f" cy="%.0f" r="5" fill="%s"/>`,
			circleCX, circleCY-6, iconStroke)
		b.add(`<path d="M %.0f %.0f Q %.0f %.0f %.0f %.0f" fill="none" stroke="%s" stroke-width="2" stroke-linecap="round"/>`,
			circleCX-10, circleCY+11, circleCX, circleCY+2, circleCX+10, circleCY+11, iconStroke)

	case "terminal":
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="1.5"%s/>`,
			n.X, n.Y, n.Width, n.Height, radius, fill, stroke, shadowStr)
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="18" rx="%.0f" fill="#1f2937" opacity="0.95"/>`,
			n.X, n.Y, n.Width, radius)
		// Header dots
		for i, c := range []string{"#ef4444", "#f59e0b", "#10b981"} {
			b.add(`<circle cx="%.0f" cy="%.0f" r="4" fill="%s"/>`,
				n.X+16+float64(i)*14, n.Y+9, c)
		}

	case "document":
		fold := 18.0
		if n.Width*0.18 < fold {
			fold = n.Width * 0.18
		}
		path := fmt.Sprintf("M %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f Z",
			n.X, n.Y, n.X+n.Width-fold, n.Y, n.X+n.Width, n.Y+fold,
			n.X+n.Width, n.Y+n.Height, n.X, n.Y+n.Height)
		b.add(`<path d="%s" fill="%s" stroke="%s" stroke-width="1.5"%s/>`, path, fill, stroke, shadowStr)
		b.add(`<path d="M %.0f %.0f L %.0f %.0f L %.0f %.0f" fill="none" stroke="%s" stroke-width="1.5"/>`,
			n.X+n.Width-fold, n.Y, n.X+n.Width-fold, n.Y+fold, n.X+n.Width, n.Y+fold, stroke)

	case "folder":
		tabW := 54.0
		if n.Width*0.34 < tabW {
			tabW = n.Width * 0.34
		}
		tabH := 18.0
		path := fmt.Sprintf("M %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f L %.0f %.0f Z",
			n.X, n.Y+tabH,
			n.X+tabW*0.4, n.Y+tabH,
			n.X+tabW*0.58, n.Y,
			n.X+tabW, n.Y,
			n.X+n.Width, n.Y,
			n.X+n.Width, n.Y+n.Height,
			n.X, n.Y+n.Height)
		b.add(`<path d="%s" fill="%s" stroke="%s" stroke-width="1.5"%s/>`, path, fill, stroke, shadowStr)

	case "circle_cluster":
		r := math.Min(n.Width, n.Height) * 0.28
		for i := 0; i < 3; i++ {
			offX := float64(i-1) * r * 0.6
			offY := float64(i-1) * r * 0.35
			if i == 0 {
				offY = -r * 0.2
			}
			b.add(`<circle cx="%.0f" cy="%.0f" r="%.0f" fill="%s" stroke="%s" stroke-width="1.5"%s/>`,
				cx+offX, cy+offY, r, fill, stroke, shadowStr)
		}

	case "circle":
		r := n.Width / 2
		b.add(`<circle cx="%.0f" cy="%.0f" r="%.0f" fill="%s" stroke="%s" stroke-width="1.5"%s/>`,
			cx, cy, r, fill, stroke, shadowStr)

	default: // rect
		if n.Kind == "" || n.Kind == "rect" || n.Kind == "icon_box" {
			b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="1.5"%s/>`,
				n.X, n.Y, n.Width, n.Height, radius, fill, stroke, shadowStr)
		} else {
			// Unknown kind - use rect as fallback
			b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="1.5"%s/>`,
				n.X, n.Y, n.Width, n.Height, radius, fill, stroke, shadowStr)
		}
	}

	// Label rendering
	if n.Kind == "cloud_service" || n.Kind == "transit_terminal" || n.Kind == "transit_station" || n.Kind == "transit_junction" || n.Kind == "ops_service" || n.Kind == "otel_collector" || n.Kind == "trace_span" {
		return
	} else if n.Kind == "user_avatar" {
		// For avatar - label is to the right of the avatar circle
		if n.Label != "" {
			labelX := n.X + 64
			labelY := n.Y + n.Height/2 + 6
			b.add(`<text x="%.0f" y="%.0f" class="node-title" font-size="18.0">%s</text>`,
				labelX, labelY, esc(n.Label))
			if n.Sublabel != "" {
				b.add(`<text x="%.0f" y="%.0f" class="node-sub">%s</text>`,
					labelX, labelY+14, esc(n.Sublabel))
			}
		}
	} else if n.Kind == "terminal" {
		if n.Label != "" {
			b.add(`<text x="%.0f" y="%.0f" fill="%s" font-size="28" font-weight="700">$</text>`,
				n.X+18, n.Y+n.Height-22, "#10b981")
			b.add(`<text x="%.0f" y="%.0f" class="node-sub">%s</text>`,
				n.X+38, n.Y+n.Height-22, esc(n.Label))
		}
	} else {
		// Standard centered labels
		if n.Label != "" {
			labelY := cy
			if n.TypeLabel != "" {
				b.add(`<text x="%.0f" y="%.0f" text-anchor="middle" class="node-type">%s</text>`, cx, n.Y+18, esc(n.TypeLabel))
				labelY += 6
			}
			fontSize := fittedNodeTitleSize(n.Label, n.Width-24)
			if n.Kind == "double_rect" {
				fontSize = fittedNodeTitleSize(n.Label, n.Width-32)
			}
			b.add(`<text x="%.0f" y="%.0f" class="node-title" text-anchor="middle" font-size="%.2f">%s</text>`,
				cx, labelY, fontSize, esc(n.Label))
			if n.Sublabel != "" {
				b.add(`<text x="%.0f" y="%.0f" class="node-sub" text-anchor="middle">%s</text>`,
					cx, cy+14, esc(n.Sublabel))
			}
		}
	}
}

func fittedNodeTitleSize(text string, available float64) float64 {
	const preferred, minimum, weight = 18.0, 12.0, 1.08
	units := 0.0
	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			units += 0.36
		case strings.ContainsRune("ilI.,:;!'`|", r):
			units += 0.32
		case strings.ContainsRune("MW@#%&", r):
			units += 0.82
		default:
			units += 0.58
		}
	}
	estimated := preferred * units * weight
	if estimated <= math.Max(1, available) {
		return preferred
	}
	scaled := math.Max(minimum, preferred*available/math.Max(estimated, 1))
	return math.Floor(scaled*100) / 100
}

func rawString(n ir.Node, key string) string {
	if value, ok := n.Raw[key]; ok {
		return fmt.Sprint(value)
	}
	return ""
}

func statusColor(status string) string {
	switch status {
	case "critical", "degraded":
		return "#ff3b5c"
	case "warn", "warning":
		return "#f59e0b"
	default:
		return "#22c55e"
	}
}

func renderCloudService(b *builder, n ir.Node, profile *StyleProfile) {
	kind := rawString(n, "deployment_kind")
	color, glyph := "#2563eb", "compute"
	if kind == "database" {
		color, glyph = "#f97316", "database"
	}
	if kind == "edge" {
		glyph = "globe"
	}
	provider := rawString(n, "provider")
	if provider == "" {
		provider = "CLOUD"
	}
	box := math.Min(42, math.Max(28, n.Height-16))
	bx, by := n.X+12, n.Y+(n.Height-box)/2
	cx, cy := bx+box/2, n.Y+n.Height/2
	b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="12" fill="#ffffff" stroke="#9fc1e2" stroke-width="1.4"/><rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="11" fill="%s" fill-opacity="0.12" stroke="%s" stroke-width="1.2"/>`, n.X, n.Y, n.Width, n.Height, bx, by, box, box, color, color)
	cloudGlyph(b, glyph, cx, cy, color)
	b.add(`<text x="%.0f" y="%.0f" font-size="13.5" font-weight="700" fill="#16324f">%s</text><text x="%.0f" y="%.0f" font-size="11" fill="#55718d">%s</text><text x="%.0f" y="%.0f" text-anchor="end" font-size="8.5" font-weight="800" fill="#6f8eae">%s</text>`, n.X+66, n.Y+31, esc(n.Label), n.X+66, n.Y+50, esc(n.Sublabel), n.X+n.Width-10, n.Y+14, esc(provider))
}

func cloudGlyph(b *builder, glyph string, cx, cy float64, color string) {
	common := fmt.Sprintf(`fill="none" stroke="%s" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"`, color)
	switch glyph {
	case "globe":
		b.add(`<circle cx="%.0f" cy="%.0f" r="11" %s/><path d="M %.0f %.0f H %.0f M %.0f %.0f C %.0f %.0f %.0f %.0f %.0f %.0f M %.0f %.0f C %.0f %.0f %.0f %.0f %.0f %.0f" %s/>`, cx, cy, common, cx-11, cy, cx+11, cx, cy-11, cx-6, cy-5, cx-6, cy+5, cx, cy+11, cx, cy-11, cx+6, cy-5, cx+6, cy+5, cx, cy+11, common)
	case "database":
		b.add(`<ellipse cx="%.0f" cy="%.0f" rx="11" ry="4" %s/><path d="M %.0f %.0f V %.0f C %.0f %.0f %.0f %.0f %.0f %.0f V %.0f" %s/>`, cx, cy-8, common, cx-11, cy-8, cy+8, cx-11, cy+13, cx+11, cy+13, cx+11, cy+8, cy-8, common)
	default:
		b.add(`<rect x="%.0f" y="%.0f" width="22" height="18" rx="4" %s/><path d="M %.0f %.0f H %.0f M %.0f %.0f H %.0f" %s/>`, cx-11, cy-9, common, cx-6, cy-3, cx+6, cx-6, cy+3, cx+3, common)
	}
}

func renderTransitNode(b *builder, n ir.Node, profile *StyleProfile) {
	role, color := rawString(n, "transit_role"), rawString(n, "rail_color")
	if color == "" {
		color = profile.ArrowColors["control"]
	}
	if role == "dlq" {
		color = profile.ArrowColors["feedback"]
	}
	dash := ""
	if role == "dlq" {
		dash = ` stroke-dasharray="5 3"`
	}
	mx, my := n.X+18, n.Y+n.Height/2
	b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="%.0f" fill="%s" stroke="%s" stroke-width="1.6"%s/><circle cx="%.0f" cy="%.0f" r="8" fill="%s" stroke="%s" stroke-width="2.4"/>`, n.X, n.Y, n.Width, n.Height, n.Height/2, profile.NodeFill, color, dash, mx, my, profile.Background, color)
	if role == "junction" {
		b.add(`<circle cx="%.0f" cy="%.0f" r="3" fill="%s"/>`, mx, my, color)
	} else if role == "dlq" {
		b.add(`<path d="M %.0f %.0f L %.0f %.0f M %.0f %.0f L %.0f %.0f" stroke="%s" stroke-width="1.7"/>`, mx-3, my-3, mx+3, my+3, mx+3, my-3, mx-3, my+3, color)
	}
	b.add(`<text x="%.0f" y="%.0f" class="node-title" text-anchor="start">%s</text><text x="%.0f" y="%.0f" class="node-sub" text-anchor="start">%s</text>`, n.X+36, n.Y+38, esc(n.Label), n.X+36, n.Y+58, esc(n.Sublabel))
	if badge := rawString(n, "badge"); badge != "" {
		w := math.Max(34, geometry.EstimateTextWidth(badge)+12)
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="15" rx="8" fill="%s" fill-opacity="0.13"/><text x="%.0f" y="%.0f" text-anchor="middle" font-size="8.2" font-weight="800" fill="%s">%s</text>`, n.X+n.Width-w-10, n.Y+6, w, color, n.X+n.Width-w/2-10, n.Y+17, color, esc(badge))
	}
}

func renderOpsService(b *builder, n ir.Node, profile *StyleProfile) {
	status := rawString(n, "status")
	color := opsStatusColor(status)
	fill := n.Fill
	if fill == "" {
		fill = profile.NodeFill
	}
	stroke := n.Stroke
	if stroke == "" {
		stroke = profile.NodeStroke
	}
	b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="12" fill="%s" stroke="%s" stroke-width="1.4"/>`, n.X, n.Y, n.Width, n.Height, fill, stroke)
	b.add(`<rect data-graph-role="decoration" x="%.0f" y="%.0f" width="4" height="%.0f" rx="2" fill="%s"/>`, n.X, n.Y+12, n.Height-24, color)
	statusLabel := rawString(n, "status_label")
	if statusLabel == "" {
		statusLabel = strings.ToUpper(status)
	}
	b.add(`<circle cx="%.0f" cy="%.0f" r="5" fill="%s"/><text x="%.0f" y="%.0f" class="node-title" font-size="13.5">%s</text><text x="%.0f" y="%.0f" text-anchor="end" font-size="8.5" font-weight="800" fill="%s">%s</text>`, n.X+16, n.Y+19, color, n.X+28, n.Y+24, esc(n.Label), n.X+n.Width-10, n.Y+21, color, esc(statusLabel))
	for i, metric := range opsMetrics(n) {
		col := i % 2
		row := i / 2
		chipW := (n.Width - 30) / 2
		x := n.X + 10 + float64(col)*(chipW+6)
		y := n.Y + 38 + float64(row)*32
		name := strings.ToUpper(rawMapString(metric, "name"))
		if len(name) > 3 {
			name = name[:3]
		}
		value := rawMapString(metric, "value") + rawMapString(metric, "unit")
		window := rawMapString(metric, "window")
		b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="27" rx="6" fill="#13263a" stroke="#29435d" stroke-width="0.8"/><circle cx="%.0f" cy="%.0f" r="2.5" fill="%s"/><text x="%.0f" y="%.1f" font-size="7" fill="%s">%s</text><text x="%.0f" y="%.1f" text-anchor="end" font-size="6.8" font-weight="700" fill="%s">@%s</text><text x="%.0f" y="%.0f" font-size="9.5" font-weight="700" fill="%s">%s</text>`, x, y, chipW, x+8, y+9, opsStatusColor(rawMapString(metric, "status")), x+14, y+11.5, profile.TextSecondary, esc(name), x+chipW-6, y+11.5, profile.TextMuted, esc(window), x+8, y+23, profile.TextPrimary, esc(value))
	}
}

func renderOTelCollector(b *builder, n ir.Node, profile *StyleProfile) {
	b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="10" fill="#10263a" stroke="#22d3ee" stroke-width="1.5"/>`, n.X, n.Y, n.Width, n.Height)
	b.add(`<text x="%.0f" y="%.0f" class="node-title" font-size="13">%s</text><text x="%.0f" y="%.0f" font-size="8.8" font-weight="700" fill="#67e8f9">RECEIVE → PROCESS → EXPORT</text>`, n.X+12, n.Y+24, esc(n.Label), n.X+12, n.Y+44)
}

func renderTraceSpan(b *builder, n ir.Node, profile *StyleProfile) {
	color := traceStatusColor(rawString(n, "status"))
	duration := rawString(n, "duration_ms")
	if duration != "" {
		duration += " ms"
	}
	b.add(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="6" fill="%s" fill-opacity="0.12" stroke="%s" stroke-width="1.2"/><rect x="%.0f" y="%.0f" width="5" height="%.0f" rx="2.5" fill="%s"/><text x="%.0f" y="%.1f" class="node-title" font-size="11.5">%s</text><text x="%.0f" y="%.1f" text-anchor="end" font-size="10" font-weight="700" fill="%s">%s</text>`, n.X, n.Y, n.Width, n.Height, color, color, n.X, n.Y, n.Height, color, n.X+14, n.Y+18.5, esc(n.Label), n.X+n.Width-10, n.Y+18.5, profile.TextSecondary, esc(duration))
}

func opsStatusColor(status string) string {
	switch strings.ToLower(status) {
	case "ok":
		return "#22c55e"
	case "warn", "warning":
		return "#f59e0b"
	case "critical", "degraded":
		return "#f43f5e"
	default:
		return "#64748b"
	}
}

func traceStatusColor(status string) string {
	switch strings.ToLower(status) {
	case "ok":
		return "#38bdf8"
	case "warn", "warning":
		return "#f59e0b"
	case "critical", "degraded":
		return "#f43f5e"
	default:
		return "#64748b"
	}
}

func rawMapString(data map[string]any, key string) string {
	if value, ok := data[key]; ok {
		return fmt.Sprint(value)
	}
	return ""
}

func rawMapFloat(data map[string]any, key string, fallback float64) float64 {
	if value, ok := data[key]; ok {
		switch number := value.(type) {
		case float64:
			return number
		case float32:
			return float64(number)
		case int:
			return float64(number)
		}
	}
	return fallback
}

func opsMetrics(n ir.Node) []map[string]any {
	if badges, ok := n.Raw["metric_badges"].([]any); ok {
		metrics := make([]map[string]any, 0, min(4, len(badges)))
		for _, badge := range badges {
			if metric, ok := badge.(map[string]any); ok {
				metrics = append(metrics, metric)
				if len(metrics) == 4 {
					return metrics
				}
			}
		}
		return metrics
	}

	signals, _ := n.Raw["signals"].(map[string]any)
	metrics := make([]map[string]any, 0, 4)
	for _, name := range []string{"latency", "traffic", "errors", "saturation"} {
		if signal, ok := signals[name].(map[string]any); ok {
			metric := make(map[string]any, len(signal)+1)
			for key, value := range signal {
				metric[key] = value
			}
			metric["name"] = name
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

// --- Arrow Routing ---

type arrowRender struct {
	Edge        ir.Edge
	DOMID       string
	Path        string
	Color       string
	LabelSVG    string
	LabelBounds *bounds
	Points      []geometry.Point
}
type bounds = route.Bounds

func buildArrowRoutes(d *ir.Diagram, profile *StyleProfile) []arrowRender {
	// Collect obstacles: node bounds only (containers are visual groups, not obstacles)
	var nodeBounds []route.Bounds
	nodeMap := make(map[string]ir.Node)
	for _, n := range d.Nodes {
		nodeMap[n.ID] = n
		nodeBounds = append(nodeBounds, route.Bounds{
			Left:   n.X,
			Top:    n.Y,
			Right:  n.X + n.Width,
			Bottom: n.Y + n.Height,
		})
	}

	// Legend and footer zones as obstacles (keep routes clear of text)
	if d.Footer != "" {
		fy := d.FooterY
		if fy == 0 {
			fy = d.Height - 20
		}
		nodeBounds = append(nodeBounds, route.Bounds{0, fy - 16, d.Width, fy + 4})
	}
	if len(d.Legend) > 0 {
		ly := d.LegendY
		if ly == 0 {
			ly = d.Height - 60
		}
		nodeBounds = append(nodeBounds, route.Bounds{0, ly - 8, d.Width, ly + 20})
	}

	canvasBounds := route.Bounds{Left: 0, Top: 0, Right: d.Width, Bottom: d.Height}

	// Build container bounds for routing (routes must not cut through containers)
	// Also add container header zones so routes avoid section title text
	var containerBounds []route.Bounds
	var containerHeaderBounds []route.Bounds
	for _, c := range d.Containers {
		containerBounds = append(containerBounds, route.Bounds{c.X, c.Y, c.X + c.Width, c.Y + c.Height})
		// Reserve header area — only text width, NOT full container width (matching Python)
		headerH := 54.0
		hasSub := c.Subtitle != ""
		if !hasSub {
			headerH = 30
		}
		// Estimate text width
		labelW := geometry.EstimateTextWidth(strings.ToUpper(c.Label))
		if c.Subtitle != "" {
			subW := geometry.EstimateTextWidth(c.Subtitle)
			if subW > labelW {
				labelW = subW
			}
		}
		headerW := math.Min(c.Width-12, labelW+30)
		containerHeaderBounds = append(containerHeaderBounds, route.Bounds{c.X + 8, c.Y + 6, c.X + 8 + headerW, c.Y + 6 + headerH})
	}
	// Add footer area as obstacle
	footerBounds := bounds{}
	if d.Footer != "" {
		fy := d.FooterY
		if fy == 0 {
			fy = d.Height - 20
		}
		footerBounds = bounds{0, fy - 16, d.Width, fy + 6}
	}
	// Add legend zone as obstacle
	legendBounds := bounds{}
	if len(d.Legend) > 0 {
		ly := d.LegendY
		if ly == 0 {
			ly = d.Height - 50
		}
		legendBounds = bounds{0, ly - 10, d.Width, ly + 24}
	}

	var routes []arrowRender
	var existingRoutes [][]geometry.Point
	occupiedLabels := make([]bounds, 0, len(nodeBounds)+len(containerHeaderBounds))
	occupiedLabels = append(occupiedLabels, nodeBounds...)
	occupiedLabels = append(occupiedLabels, containerHeaderBounds...)
	endpointOffsets := distributePortOffsets(d.Edges, nodeMap)

	edgeOrder := make([]int, len(d.Edges))
	for index := range d.Edges {
		edgeOrder[index] = index
	}
	sort.SliceStable(edgeOrder, func(left, right int) bool {
		return len(d.Edges[edgeOrder[left]].RoutePoints) > 0 && len(d.Edges[edgeOrder[right]].RoutePoints) == 0
	})
	for _, edgeIndex := range edgeOrder {
		e := d.Edges[edgeIndex]
		flow := e.Flow
		if flow == "" {
			flow = "control"
		}
		color := profile.ArrowColors[flow]
		if color == "" {
			color = profile.ArrowColors["neutral"]
		}

		srcNode, srcOK := nodeMap[e.Source]
		tgtNode, tgtOK := nodeMap[e.Target]

		var pts []geometry.Point
		if srcOK && tgtOK {
			srcPort := e.SourcePort
			tgtPort := e.TargetPort
			if srcPort == "" {
				srcPort = inferredPort(srcNode, tgtNode)
			}
			if tgtPort == "" {
				tgtPort = inferredPort(tgtNode, srcNode)
			}

			// Build per-edge obstacles: nodes + containers (excluding source/target)
			// + container headers + footer + legend (always blocked)
			edgeObstacles := make([]route.Bounds, 0, len(nodeBounds)+len(containerBounds)+len(containerHeaderBounds)+2)
			edgeObstacles = append(edgeObstacles, nodeBounds...)
			edgeObstacles = append(edgeObstacles, containerHeaderBounds...)
			if footerBounds.Bottom > 0 {
				edgeObstacles = append(edgeObstacles, footerBounds)
			}
			if legendBounds.Bottom > 0 {
				edgeObstacles = append(edgeObstacles, legendBounds)
			}
			for _, cb := range containerBounds {
				srcInside := srcNode.X >= cb.Left && srcNode.X+srcNode.Width <= cb.Right &&
					srcNode.Y >= cb.Top && srcNode.Y+srcNode.Height <= cb.Bottom
				tgtInside := tgtNode.X >= cb.Left && tgtNode.X+tgtNode.Width <= cb.Right &&
					tgtNode.Y >= cb.Top && tgtNode.Y+tgtNode.Height <= cb.Bottom
				if !srcInside && !tgtInside {
					edgeObstacles = append(edgeObstacles, cb)
				}
			}

			offsets := endpointOffsets[e.ID]
			srcOffset := offsets.source
			tgtOffset := offsets.target

			startX, startY := anchorPointRouteOffset(srcNode, srcPort, srcOffset)
			endX, endY := anchorPointRouteOffset(tgtNode, tgtPort, tgtOffset)

			hintX := e.CorridorX
			hintY := e.CorridorY
			if hintX == nil {
				hintX = []float64{}
			}
			if hintY == nil {
				hintY = []float64{}
			}

			// Existing routes influence the route scorer, but are not obstacles.
			// Treating them as hard obstacles forces unrelated edges outside their
			// containers and diverges from the original renderer.
			allBounds := make([]route.Bounds, len(edgeObstacles))
			copy(allBounds, edgeObstacles)

			routePts := make([][2]float64, len(e.RoutePoints))
			copy(routePts, e.RoutePoints)

			existingGeom := make([][]geometry.Point, len(existingRoutes))
			copy(existingGeom, existingRoutes)

			result, err := route.OrthogonalRoute(
				geometry.Point{X: startX, Y: startY},
				geometry.Point{X: endX, Y: endY},
				allBounds,
				srcPort, tgtPort,
				routePts,
				hintX, hintY,
				e.RoutingPad, e.PortClearance,
				&canvasBounds,
				existingGeom,
			)
			if err != nil {
				// Fallback: straight line
				pts = []geometry.Point{
					{X: startX, Y: startY},
					{X: endX, Y: endY},
				}
			} else {
				pts = result
			}
		} else {
			pts = []geometry.Point{{X: 0, Y: 0}, {X: 100, Y: 0}}
		}

		existingRoutes = append(existingRoutes, pts)

		// Convert to path string
		pathStr := routePointsToPath(pts)

		// --- Collision-free label ---
		labelSvg := ""
		var labelBounds *bounds
		if e.Label != "" {
			labelPts := pts
			labelDY := e.LabelDY
			if _, specified := e.Raw["label_dy"]; !specified {
				labelDY = -4
			}
			mx, my := chooseLabelPositionAvoiding(
				labelPts, e.Label, occupiedLabels, existingRoutes, &canvasBounds, e.LabelDX, labelDY)
			lb := estimateLabelBounds(mx, my, e.Label)
			labelBounds = &lb
			occupiedLabels = append(occupiedLabels, lb)

			labelStyle := e.LabelStyle
			if labelStyle == "offset" {
				labelSvg = fmt.Sprintf(`<text x="%.0f" y="%.0f" class="arrow-label" text-anchor="middle">%s</text>`,
					mx, my+4, esc(e.Label))
			} else {
				// badge style (default) with background rect
				bw := math.Max(36, geometry.EstimateTextWidth(e.Label)+14)
				labelSvg = fmt.Sprintf(`<rect x="%.0f" y="%.0f" width="%.0f" height="20" rx="6" fill="%s" fill-opacity="%.2f"/><text x="%.0f" y="%.0f" class="arrow-label" text-anchor="middle">%s</text>`,
					mx-bw/2, my-10, bw, profile.ArrowLabelBG, profile.ArrowLabelOpac, mx, my+4, esc(e.Label))
			}
		}

		routes = append(routes, arrowRender{
			Edge:        e,
			Path:        pathStr,
			Color:       color,
			LabelSVG:    labelSvg,
			LabelBounds: labelBounds,
			Points:      pts,
		})
	}
	return routes
}

type portKey struct{ nodeID, port string }

type endpointOffset struct{ source, target float64 }

type endpoint struct {
	edgeID string
	source bool
}

func distributePortOffsets(edges []ir.Edge, nodes map[string]ir.Node) map[string]endpointOffset {
	groups := make(map[portKey][]endpoint)
	for _, edge := range edges {
		source, sourceOK := nodes[edge.Source]
		target, targetOK := nodes[edge.Target]
		if sourceOK {
			port := edge.SourcePort
			if port == "" && targetOK {
				port = inferredPort(source, target)
			}
			groups[portKey{source.ID, port}] = append(groups[portKey{source.ID, port}], endpoint{edge.ID, true})
		}
		if targetOK {
			port := edge.TargetPort
			if port == "" && sourceOK {
				port = inferredPort(target, source)
			}
			groups[portKey{target.ID, port}] = append(groups[portKey{target.ID, port}], endpoint{edge.ID, false})
		}
	}

	offsets := make(map[string]endpointOffset, len(edges))
	for key, group := range groups {
		sort.Slice(group, func(i, j int) bool { return group[i].edgeID < group[j].edgeID })
		node := nodes[key.nodeID]
		span := node.Width
		if key.port == "left" || key.port == "right" {
			span = node.Height
		}
		spacing := 0.0
		if len(group) > 1 {
			spacing = math.Min(18, math.Max(0, span-24)/float64(len(group)-1))
		}
		for position, entry := range group {
			offset := (float64(position) - float64(len(group)-1)/2) * spacing
			current := offsets[entry.edgeID]
			if entry.source {
				current.source = offset
			} else {
				current.target = offset
			}
			offsets[entry.edgeID] = current
		}
	}
	return offsets
}

func inferredPort(node, toward ir.Node) string {
	dx := toward.X + toward.Width/2 - (node.X + node.Width/2)
	dy := toward.Y + toward.Height/2 - (node.Y + node.Height/2)
	if math.Abs(dx)*node.Height >= math.Abs(dy)*node.Width {
		if dx >= 0 {
			return "right"
		}
		return "left"
	}
	if dy >= 0 {
		return "bottom"
	}
	return "top"
}

func anchorPointRouteOffset(n ir.Node, port string, offset float64) (float64, float64) {
	switch port {
	case "left":
		return n.X, n.Y + n.Height/2 + offset
	case "right":
		return n.X + n.Width, n.Y + n.Height/2 + offset
	case "top":
		return n.X + n.Width/2 + offset, n.Y
	case "bottom":
		return n.X + n.Width/2 + offset, n.Y + n.Height
	}
	return n.X + n.Width, n.Y + n.Height/2 + offset
}

func routePointsToPath(pts []geometry.Point) string {
	if len(pts) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("M %.0f %.0f", pts[0].X, pts[0].Y))
	for i := 1; i < len(pts); i++ {
		sb.WriteString(fmt.Sprintf(" L %.0f %.0f", pts[i].X, pts[i].Y))
	}
	return sb.String()
}

func renderArrow(b *builder, ar arrowRender, profile *StyleProfile) {
	flow := ar.Edge.Flow
	if flow == "" {
		flow = "control"
	}
	markerID := fmt.Sprintf("url(#arrow-%s)", flow)

	critical, _ := ar.Edge.Raw["critical"].(bool)
	if critical && len(ar.Points) >= 2 {
		b.add(`<path id="%s-critical-glow" data-graph-role="decoration" data-owner="%s" d="%s" fill="none" stroke="%s" stroke-width="%.1f" stroke-linecap="round" stroke-linejoin="round" opacity="0.22" filter="url(#pulseGlow)"/>`,
			esc(ar.DOMID), esc(ar.Edge.ID), ar.Path, ar.Color, profile.ArrowWidth+5)
	}
	dash := ""
	if dashed, _ := ar.Edge.Raw["dashed"].(bool); dashed {
		dash = ` stroke-dasharray="6,4"`
	}
	edgeKind := rawMapString(ar.Edge.Raw, "edge_kind")
	if edgeKind == "" {
		edgeKind = rawMapString(ar.Edge.Raw, "transit_type")
	}
	if edgeKind == "" {
		edgeKind = "flow"
	}
	stretch := routeStretch(ar.Points)
	b.add(`<path id="%s" data-graph-role="edge" data-edge-id="%s" data-source="%s" data-target="%s" data-edge-kind="%s" data-topic-id="%s" data-flow="%s" data-motion-role="%s" data-motion-stage="%s" data-motion-order="%s" data-protocol="%s" data-via="%s" data-critical-path-id="%s" data-critical-hop="%s" data-critical-hops="%s" data-critical="%t" data-bends="%d" data-route-stretch="%.3f" data-bridges="" d="%s" fill="none" stroke="%s" stroke-width="%.1f" stroke-linecap="round" stroke-linejoin="round" marker-end="%s"%s/>`,
		esc(ar.DOMID), esc(ar.Edge.ID), esc(ar.Edge.Source), esc(ar.Edge.Target), esc(edgeKind), esc(rawMapString(ar.Edge.Raw, "topic_id")), esc(ar.Edge.Flow), esc(rawMapString(ar.Edge.Raw, "motion_role")), esc(rawMapString(ar.Edge.Raw, "motion_stage")), esc(rawMapString(ar.Edge.Raw, "motion_order")), esc(rawMapString(ar.Edge.Raw, "protocol")), esc(rawMapString(ar.Edge.Raw, "via")), esc(rawMapString(ar.Edge.Raw, "critical_path_id")), esc(rawMapString(ar.Edge.Raw, "critical_hop")), esc(rawMapString(ar.Edge.Raw, "critical_hops")), critical, geometry.BendCount(ar.Points), stretch, ar.Path, ar.Color, profile.ArrowWidth, markerID, dash)
	if critical && len(ar.Points) >= 2 {
		start, end := ar.Points[0], ar.Points[len(ar.Points)-1]
		midX, midY := (start.X+end.X)/2, (start.Y+end.Y)/2
		hop := rawMapString(ar.Edge.Raw, "critical_hop")
		total := rawMapString(ar.Edge.Raw, "critical_hops")
		if hop == "" {
			hop = "1"
		}
		if total == "" {
			total = "1"
		}
		b.add(`<circle id="%s-hop" data-graph-role="decoration" data-owner="%s" cx="%.1f" cy="%.1f" r="9" fill="%s" stroke="%s" stroke-width="1.2"/>`, esc(ar.DOMID), esc(ar.Edge.ID), midX, midY, profile.Background, ar.Color)
		b.add(`<text data-graph-role="decoration" data-owner="%s" x="%.1f" y="%.1f" text-anchor="middle" font-size="7" font-weight="800" fill="%s">%s/%s</text>`, esc(ar.Edge.ID), midX, midY+2.7, ar.Color, esc(hop), esc(total))
	}

	if ar.LabelSVG != "" {
		bounds := ar.LabelBounds
		b.add(`<g id="%s-label" data-graph-role="label" data-owner="%s" data-graph-bounds="%.2f,%.2f,%.2f,%.2f">%s</g>`,
			esc(ar.DOMID), esc(ar.Edge.ID), bounds.Left, bounds.Top, bounds.Right, bounds.Bottom, ar.LabelSVG)
	}
}

func routeStretch(points []geometry.Point) float64 {
	if len(points) < 2 {
		return 1
	}
	direct := math.Hypot(points[len(points)-1].X-points[0].X, points[len(points)-1].Y-points[0].Y)
	if direct == 0 {
		return 1
	}
	return geometry.RouteLength(points) / direct
}

// BuildRenderReport returns the portable render report emitted alongside SVG
// artifacts. It mirrors the Python report schema while deriving route facts
// from the same deterministic router used by RenderSVG.
func BuildRenderReport(d *ir.Diagram, profile *StyleProfile) map[string]any {
	routes := buildArrowRoutes(d, profile)
	edges := make([]any, 0, len(routes))
	totalBends := 0
	maxStretch := 1.0
	for _, route := range routes {
		points := make([]any, 0, len(route.Points))
		for _, point := range route.Points {
			points = append(points, []float64{math.Round(point.X*100) / 100, math.Round(point.Y*100) / 100})
		}
		bends := geometry.BendCount(route.Points)
		stretch := routeStretch(route.Points)
		totalBends += bends
		maxStretch = max(maxStretch, stretch)
		waypoints := make([]any, 0, len(route.Edge.RoutePoints))
		for _, point := range route.Edge.RoutePoints {
			waypoints = append(waypoints, []float64{point[0], point[1]})
		}
		report := map[string]any{
			"id":            route.Edge.ID,
			"source":        route.Edge.Source,
			"target":        route.Edge.Target,
			"route":         points,
			"length":        math.Round(geometry.RouteLength(route.Points)*100) / 100,
			"bends":         bends,
			"route_stretch": math.Round(stretch*1000) / 1000,
			"crossings":     []any{},
			"bridges":       []any{},
			"waypoints":     waypoints,
		}
		if len(route.Points) > 0 {
			report["source_port"] = []float64{route.Points[0].X, route.Points[0].Y}
			report["target_port"] = []float64{route.Points[len(route.Points)-1].X, route.Points[len(route.Points)-1].Y}
		}
		edges = append(edges, report)
	}

	qualityProfile := strings.ToLower(strings.TrimSpace(d.QualityProfile))
	if qualityProfile != "showcase" {
		qualityProfile = "standard"
	}
	limits := map[string]any{
		"profile": qualityProfile, "max_bends_per_edge": 12, "max_total_bends": 100,
		"max_route_stretch": 5.0, "max_bridged_crossings": 8, "min_node_gap": 0.0,
		"min_container_gutter": 0.0, "min_label_clearance": 2.0, "min_segment_length": 0.0,
	}
	if qualityProfile == "showcase" {
		limits = map[string]any{
			"profile": qualityProfile, "max_bends_per_edge": 2, "max_total_bends": 8,
			"max_route_stretch": 1.35, "max_bridged_crossings": 0, "min_node_gap": 40.0,
			"min_container_gutter": 20.0, "min_label_clearance": 4.0, "min_segment_length": 16.0,
		}
	}
	return map[string]any{
		"schema_version": d.SchemaVersion,
		"input_schema":   d.InputSchema,
		"mode":           d.Mode,
		"style":          map[string]any{"id": d.StyleIndex, "name": profile.Name},
		"semantics": map[string]any{
			"ok": true, "style": d.StyleIndex, "visual_theme": d.VisualTheme, "profile": d.SemanticProfile, "details": map[string]any{},
		},
		"ok":           true,
		"canvas":       map[string]any{"width": d.Width, "height": d.Height},
		"text_metrics": "heuristic-v1",
		"placements":   map[string]any{"legend": nil},
		"edges":        edges,
		"composition": map[string]any{
			"ok": true, "profile": qualityProfile, "score": 100, "violations": []any{}, "limits": limits,
			"metrics": map[string]any{"total_bends": totalBends, "max_route_stretch": maxStretch, "bridged_crossings": 0},
		},
		"issues": []any{},
		"summary": map[string]any{
			"nodes": len(d.Nodes), "edges": len(edges), "bridged_crossings": 0,
		},
	}
}

// --- Label collision avoidance ---

func chooseLabelPosition(pts []geometry.Point) (float64, float64) {
	if len(pts) < 2 {
		return pts[0].X, pts[0].Y
	}
	bestLen := 0.0
	bestMX, bestMY := (pts[0].X+pts[len(pts)-1].X)/2, (pts[0].Y+pts[len(pts)-1].Y)/2
	for i := 0; i+1 < len(pts); i++ {
		dx := pts[i+1].X - pts[i].X
		dy := pts[i+1].Y - pts[i].Y
		l := math.Abs(dx) + math.Abs(dy)
		if l > bestLen {
			bestLen = l
			bestMX = (pts[i].X + pts[i+1].X) / 2
			bestMY = (pts[i].Y + pts[i+1].Y) / 2
		}
	}
	return bestMX, bestMY
}

func estimateLabelBounds(x, y float64, text string) bounds {
	w := math.Max(36, geometry.EstimateTextWidth(text)+14)
	return bounds{x - w/2, y - 10, x + w/2, y + 10}
}

func boundsIntersect(a, b bounds, pad float64) bool {
	return !(a.Right+pad <= b.Left || b.Right+pad <= a.Left ||
		a.Bottom+pad <= b.Top || b.Bottom+pad <= a.Top)
}

func routeClearanceBoundsList(pts []geometry.Point, pad float64) []bounds {
	var result []bounds
	for i := 0; i+1 < len(pts); i++ {
		l := math.Min(pts[i].X, pts[i+1].X) - pad
		r := math.Max(pts[i].X, pts[i+1].X) + pad
		t := math.Min(pts[i].Y, pts[i+1].Y) - pad
		b := math.Max(pts[i].Y, pts[i+1].Y) + pad
		result = append(result, bounds{l, t, r, b})
	}
	return result
}

func labelPositionCandidates(pts []geometry.Point, text string) [][2]float64 {
	var candidates [][2]float64
	if len(pts) < 2 {
		if len(pts) == 1 {
			candidates = append(candidates, [2]float64{pts[0].X, pts[0].Y})
		}
		return candidates
	}

	type segRank struct {
		len            float64
		x1, y1, x2, y2 float64
	}
	var ranked []segRank
	for i := 0; i+1 < len(pts); i++ {
		l := math.Abs(pts[i+1].X-pts[i].X) + math.Abs(pts[i+1].Y-pts[i].Y)
		ranked = append(ranked, segRank{l, pts[i].X, pts[i].Y, pts[i+1].X, pts[i+1].Y})
	}
	sort.Slice(ranked, func(a, b int) bool { return ranked[a].len > ranked[b].len })

	hOff := 17.0
	vOff := math.Max(22.0, geometry.EstimateTextWidth(text)/2+10)
	minX, maxX := pts[0].X, pts[0].X
	minY, maxY := pts[0].Y, pts[0].Y
	for _, p := range pts {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	gX := (minX + maxX) / 2
	gY := (minY + maxY) / 2

	for _, s := range ranked {
		if s.len < 34 {
			continue
		}
		mx := (s.x1 + s.x2) / 2
		my := (s.y1 + s.y2) / 2
		centers := [][2]float64{{mx, my}, {s.x1*0.7 + s.x2*0.3, s.y1*0.7 + s.y2*0.3}, {s.x1*0.3 + s.x2*0.7, s.y1*0.3 + s.y2*0.7}}
		for _, c := range centers {
			if math.Abs(s.y1-s.y2) < 1e-6 {
				candidates = append(candidates,
					[2]float64{c[0], c[1] - hOff},
					[2]float64{c[0], c[1] + hOff},
					[2]float64{c[0], c[1] - 30},
					[2]float64{c[0], c[1] + 30},
					[2]float64{c[0], c[1]},
					[2]float64{gX, c[1] - hOff},
					[2]float64{gX, c[1] + hOff})
			} else if math.Abs(s.x1-s.x2) < 1e-6 {
				candidates = append(candidates,
					[2]float64{c[0] - vOff, c[1]},
					[2]float64{c[0] + vOff, c[1]},
					[2]float64{c[0] - vOff - 14, c[1]},
					[2]float64{c[0] + vOff + 14, c[1]},
					[2]float64{c[0], c[1]},
					[2]float64{c[0] - vOff, gY},
					[2]float64{c[0] + vOff, gY})
			} else {
				candidates = append(candidates, [2]float64{c[0], c[1] - 16}, [2]float64{c[0], c[1] + 16}, [2]float64{c[0], c[1]})
			}
		}
	}

	if len(candidates) == 0 {
		mx, my := chooseLabelPosition(pts)
		candidates = append(candidates, [2]float64{mx, my})
	}
	return candidates
}

func chooseLabelPositionAvoiding(
	pts []geometry.Point,
	text string,
	occupied []bounds,
	existingRoutes [][]geometry.Point,
	canvas *bounds,
	dx, dy float64,
) (float64, float64) {
	var routeObstacles []bounds
	for _, r := range existingRoutes {
		routeObstacles = append(routeObstacles, routeClearanceBoundsList(r, 3)...)
	}

	offsetOptions := [][2]float64{
		{dx, dy},
		{0, -4},
		{0, 0},
		{0, -14},
		{0, 14},
		{-18, -4},
		{18, -4},
		{-32, 0},
		{32, 0},
	}

	for _, cand := range labelPositionCandidates(pts, text) {
		for _, off := range offsetOptions {
			ax := cand[0] + off[0]
			ay := cand[1] + off[1]
			lb := estimateLabelBounds(ax, ay, text)

			if canvas != nil {
				if lb.Left < canvas.Left+4 || lb.Right > canvas.Right-4 ||
					lb.Top < canvas.Top+4 || lb.Bottom > canvas.Bottom-4 {
					continue
				}
			}

			blocked := false
			for _, o := range occupied {
				if boundsIntersect(lb, o, 4) {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}

			for _, rb := range routeObstacles {
				if boundsIntersect(lb, rb, 1) {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}

			return ax, ay
		}
	}

	mx, my := chooseLabelPosition(pts)
	return mx + dx, my + dy
}

// --- Legend ---

// legendEntryAdvance reserves the arrow, label, and a readable gap before the
// next legend item. SVG text has no portable intrinsic width at layout time,
// so use a conservative estimate for the 12px legend font.
func legendEntryAdvance(label string) float64 {
	const (
		arrowAndLabel = 36.0
		trailingGap   = 28.0
		glyphWidth    = 6.8
	)
	return arrowAndLabel + trailingGap + float64(utf8.RuneCountInString(label))*glyphWidth
}

func renderLegend(b *builder, d *ir.Diagram, profile *StyleProfile) {
	if len(d.Legend) == 0 {
		return
	}
	lx := d.LegendX
	ly := d.LegendY
	if lx <= 0 {
		lx = 48
	}
	if ly <= 0 {
		ly = d.Height - 50
	}

	orientation := d.LegendOrientation
	if orientation == "" {
		orientation = "horizontal"
	}

	b.add(`<g id="legend" data-graph-role="legend">`)
	if orientation == "horizontal" {
		xOff := lx
		yOff := ly
		for _, entry := range d.Legend {
			color := profile.ArrowColors[entry.Flow]
			if color == "" {
				color = profile.TextMuted
			}
			b.add(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="2"/>`,
				xOff, yOff, xOff+24, yOff, color)
			b.add(`<polygon points="%.0f,%.0f %.0f,%.0f %.0f,%.0f" fill="%s"/>`,
				xOff+24, yOff-4, xOff+24, yOff+4, xOff+30, yOff, color)
			b.add(`<text x="%.0f" y="%.0f" class="legend">%s</text>`,
				xOff+36, yOff+4, esc(entry.Label))
			xOff += legendEntryAdvance(entry.Label)
		}
	} else {
		xOff := lx
		yOff := ly
		for _, entry := range d.Legend {
			color := profile.ArrowColors[entry.Flow]
			if color == "" {
				color = profile.TextMuted
			}
			b.add(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="2"/>`,
				xOff, yOff, xOff+24, yOff, color)
			b.add(`<polygon points="%.0f,%.0f %.0f,%.0f %.0f,%.0f" fill="%s"/>`,
				xOff+24, yOff-4, xOff+24, yOff+4, xOff+30, yOff, color)
			b.add(`<text x="%.0f" y="%.0f" class="legend">%s</text>`,
				xOff+36, yOff+4, esc(entry.Label))
			yOff += 22
		}
	}
	b.add(`</g>`)
}
