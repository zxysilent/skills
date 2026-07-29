// Package ir provides typed diagram data structures and input normalization.
// Ported from diagram_ir.py (201 lines).
package ir

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// DiagramValidationError is raised when input cannot be normalized.
type DiagramValidationError struct {
	Message string
}

func (e *DiagramValidationError) Error() string {
	return e.Message
}

func diagramError(format string, args ...any) *DiagramValidationError {
	return &DiagramValidationError{Message: fmt.Sprintf(format, args...)}
}

// Node describes one diagram node/component.
type Node struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind,omitempty"`
	X         float64        `json:"x"`
	Y         float64        `json:"y"`
	Width     float64        `json:"width"`
	Height    float64        `json:"height"`
	R         float64        `json:"r,omitempty"`
	OffsetY   float64        `json:"offset_y,omitempty"`
	Label     string         `json:"label,omitempty"`
	TypeLabel string         `json:"type_label,omitempty"`
	Sublabel  string         `json:"sublabel,omitempty"`
	Fill      string         `json:"fill,omitempty"`
	Stroke    string         `json:"stroke,omitempty"`
	Flat      bool           `json:"flat,omitempty"`
	Raw       map[string]any `json:"-"`
}

// Edge describes one arrow/edge between nodes.
type Edge struct {
	ID            string         `json:"id"`
	Source        string         `json:"source"`
	Target        string         `json:"target"`
	SourcePort    string         `json:"source_port,omitempty"`
	TargetPort    string         `json:"target_port,omitempty"`
	Flow          string         `json:"flow,omitempty"`
	Label         string         `json:"label,omitempty"`
	LabelDX       float64        `json:"label_dx,omitempty"`
	LabelDY       float64        `json:"label_dy,omitempty"`
	CorridorX     []float64      `json:"corridor_x,omitempty"`
	CorridorY     []float64      `json:"corridor_y,omitempty"`
	RoutePoints   [][2]float64   `json:"route_points,omitempty"`
	RoutingPad    float64        `json:"routing_padding,omitempty"`
	PortClearance float64        `json:"port_clearance,omitempty"`
	LabelStyle    string         `json:"label_style,omitempty"`
	MotionRole    string         `json:"motion_role,omitempty"`
	MotionStage   int            `json:"motion_stage,omitempty"`
	MotionOrder   int            `json:"motion_order,omitempty"`
	Raw           map[string]any `json:"-"`
}

// Container groups related diagram nodes.
type Container struct {
	ID             string         `json:"id"`
	X              float64        `json:"x"`
	Y              float64        `json:"y"`
	Width          float64        `json:"width"`
	Height         float64        `json:"height"`
	Label          string         `json:"label,omitempty"`
	Subtitle       string         `json:"subtitle,omitempty"`
	DeploymentKind string         `json:"deployment_kind,omitempty"`
	Raw            map[string]any `json:"-"`
}

// LegendEntry describes one entry in the arrow legend.
type LegendEntry struct {
	Flow  string `json:"flow"`
	Label string `json:"label"`
}

// Diagram is the normalized in-memory representation.
type Diagram struct {
	SchemaVersion     int            `json:"schema_version"`
	InputSchema       string         `json:"input_schema"`
	Mode              string         `json:"mode"`
	StyleIndex        int            `json:"style_index"`
	SemanticProfile   string         `json:"semantic_profile"`
	VisualTheme       string         `json:"visual_theme"`
	Width             float64        `json:"width"`
	Height            float64        `json:"height"`
	Title             string         `json:"title,omitempty"`
	Subtitle          string         `json:"subtitle,omitempty"`
	Footer            string         `json:"footer,omitempty"`
	FooterX           float64        `json:"footer_x,omitempty"`
	FooterY           float64        `json:"footer_y,omitempty"`
	MotionScene       string         `json:"motion_scene,omitempty"`
	QualityProfile    string         `json:"quality_profile,omitempty"`
	MetaLeft          string         `json:"meta_left,omitempty"`
	MetaCenter        string         `json:"meta_center,omitempty"`
	MetaRight         string         `json:"meta_right,omitempty"`
	C4Level           string         `json:"c4_level,omitempty"`
	ReviewState       string         `json:"review_state,omitempty"`
	PlatformProfile   string         `json:"platform_profile,omitempty"`
	ObservationWindow string         `json:"observation_window,omitempty"`
	Topics            []any          `json:"topics,omitempty"`
	Containers        []Container    `json:"containers"`
	Nodes             []Node         `json:"nodes"`
	Edges             []Edge         `json:"arrows"`
	Legend            []LegendEntry  `json:"legend,omitempty"`
	LegendOrientation string         `json:"legend_orientation,omitempty"`
	LegendX           float64        `json:"legend_x,omitempty"`
	LegendY           float64        `json:"legend_y,omitempty"`
	LegendLocked      bool           `json:"legend_locked,omitempty"`
	Raw               map[string]any `json:"-"`
}

func toFloat(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func finite(v any, path string) (float64, error) {
	f, err := toFloat(v)
	if err != nil {
		return 0, diagramError("%s must be a finite number", path)
	}
	if !math.IsInf(f, 0) && !math.IsNaN(f) {
		return f, nil
	}
	return 0, diagramError("%s must be a finite number", path)
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asList(v any) []any {
	list, _ := v.([]any)
	return list
}

// Normalize validates and normalizes a raw JSON map into a Diagram.
func Normalize(data map[string]any) (*Diagram, error) {
	if data == nil {
		return nil, diagramError("diagram input must be a JSON object")
	}

	d := &Diagram{
		SchemaVersion: 1,
		InputSchema:   "legacy",
		Raw:           data,
	}

	if sv, ok := data["schema_version"]; ok {
		d.InputSchema = "v1"
		var svInt int
		switch v := sv.(type) {
		case int:
			svInt = v
		case float64:
			svInt = int(v)
		default:
			return nil, diagramError("unsupported schema_version")
		}
		if svInt != 1 {
			return nil, diagramError("unsupported schema_version: %v", sv)
		}
	}

	// mode / template_type
	mode, _ := data["mode"].(string)
	templateType, _ := data["template_type"].(string)
	if mode == "" {
		mode = templateType
	}
	if mode == "" {
		mode = "architecture"
	}
	d.Mode = mode

	// dimensions
	if w, err := finite(data["width"], "diagram.width"); err == nil {
		d.Width = w
	} else {
		d.Width = 960
	}
	if h, err := finite(data["height"], "diagram.height"); err == nil {
		d.Height = h
	} else {
		d.Height = 600
	}

	// title / subtitle / footer / meta
	d.Title = asString(data["title"])
	d.Subtitle = asString(data["subtitle"])
	d.Footer = asString(data["footer"])
	if fx, err := finite(data["footer_x"], "diagram.footer_x"); err == nil {
		d.FooterX = fx
	}
	if fy, err := finite(data["footer_y"], "diagram.footer_y"); err == nil {
		d.FooterY = fy
	}
	d.MotionScene = asString(data["motion_scene"])
	d.QualityProfile = asString(data["quality_profile"])
	d.MetaLeft = asString(data["meta_left"])
	d.MetaCenter = asString(data["meta_center"])
	d.MetaRight = asString(data["meta_right"])
	d.C4Level = asString(data["c4_level"])
	d.ReviewState = asString(data["review_state"])
	d.PlatformProfile = asString(data["platform_profile"])
	d.ObservationWindow = asString(data["observation_window"])
	if t, ok := data["topics"].([]any); ok {
		d.Topics = t
	}
	d.LegendOrientation = asString(data["legend_orientation"])
	if lx, err := finite(data["legend_x"], "diagram.legend_x"); err == nil {
		d.LegendX = lx
	}
	if ly, err := finite(data["legend_y"], "diagram.legend_y"); err == nil {
		d.LegendY = ly
	}
	d.LegendLocked, _ = data["legend_locked"].(bool)

	// containers
	usedIDs := make(map[string]struct{})
	for _, raw := range asList(data["containers"]) {
		m := asMap(raw)
		if m == nil {
			return nil, diagramError("containers entry must be an object")
		}
		c := Container{Raw: m}
		c.ID = strings.TrimSpace(asString(m["id"]))
		if c.ID == "" {
			return nil, diagramError("container id must be non-empty")
		}
		if _, dup := usedIDs[c.ID]; dup {
			return nil, diagramError("duplicate container id: %s", c.ID)
		}
		usedIDs[c.ID] = struct{}{}
		var err error
		if c.X, err = finite(m["x"], fmt.Sprintf("containers[%s].x", c.ID)); err != nil {
			return nil, err
		}
		if c.Y, err = finite(m["y"], fmt.Sprintf("containers[%s].y", c.ID)); err != nil {
			return nil, err
		}
		if c.Width, err = finite(m["width"], fmt.Sprintf("containers[%s].width", c.ID)); err != nil {
			return nil, err
		}
		if c.Height, err = finite(m["height"], fmt.Sprintf("containers[%s].height", c.ID)); err != nil {
			return nil, err
		}
		c.Label = asString(m["label"])
		c.Subtitle = asString(m["subtitle"])
		c.DeploymentKind = asString(m["deployment_kind"])
		d.Containers = append(d.Containers, c)
	}

	// nodes
	for i, raw := range asList(data["nodes"]) {
		m := asMap(raw)
		if m == nil {
			return nil, diagramError("nodes[%d] must be an object", i)
		}
		n := Node{Raw: m}
		n.ID = strings.TrimSpace(asString(m["id"]))
		if n.ID == "" {
			n.ID = fmt.Sprintf("node-%03d", i)
		}
		if _, dup := usedIDs[n.ID]; dup {
			return nil, diagramError("duplicate diagram id: %s", n.ID)
		}
		usedIDs[n.ID] = struct{}{}
		var err error
		if n.X, err = finite(m["x"], fmt.Sprintf("nodes[%s].x", n.ID)); err != nil {
			return nil, err
		}
		if n.Y, err = finite(m["y"], fmt.Sprintf("nodes[%s].y", n.ID)); err != nil {
			return nil, err
		}
		if n.Width, err = finite(m["width"], fmt.Sprintf("nodes[%s].width", n.ID)); err != nil {
			return nil, err
		}
		if n.Height, err = finite(m["height"], fmt.Sprintf("nodes[%s].height", n.ID)); err != nil {
			return nil, err
		}
		n.Kind = asString(m["kind"])
		n.Label = asString(m["label"])
		n.TypeLabel = asString(m["type_label"])
		n.Sublabel = asString(m["sublabel"])
		n.Fill = asString(m["fill"])
		n.Stroke = asString(m["stroke"])
		n.Flat, _ = m["flat"].(bool)
		if r, err := finite(m["r"], fmt.Sprintf("nodes[%s].r", n.ID)); err == nil {
			n.R = r
		}
		if oy, err := finite(m["offset_y"], fmt.Sprintf("nodes[%s].offset_y", n.ID)); err == nil {
			n.OffsetY = oy
		}
		d.Nodes = append(d.Nodes, n)
	}

	// edges (arrows)
	rawEdges := asList(data["arrows"])
	if len(rawEdges) == 0 {
		if edges := asList(data["edges"]); len(edges) > 0 {
			rawEdges = edges
		}
	}
	for i, raw := range rawEdges {
		m := asMap(raw)
		if m == nil {
			return nil, diagramError("arrows[%d] must be an object", i)
		}
		e := Edge{Raw: m}
		e.ID = strings.TrimSpace(asString(m["id"]))
		if e.ID == "" {
			e.ID = fmt.Sprintf("edge-%03d", i)
		}
		if _, dup := usedIDs[e.ID]; dup {
			return nil, diagramError("duplicate diagram id: %s", e.ID)
		}
		usedIDs[e.ID] = struct{}{}
		e.Source = strings.TrimSpace(asString(m["source"]))
		e.Target = strings.TrimSpace(asString(m["target"]))
		e.SourcePort = asString(m["source_port"])
		e.TargetPort = asString(m["target_port"])
		e.Flow = asString(m["flow"])
		e.Label = asString(m["label"])
		e.LabelStyle = asString(m["label_style"])
		e.MotionRole = asString(m["motion_role"])

		if dx, err := finite(m["label_dx"], fmt.Sprintf("arrows[%s].label_dx", e.ID)); err == nil {
			e.LabelDX = dx
		}
		if dy, err := finite(m["label_dy"], fmt.Sprintf("arrows[%s].label_dy", e.ID)); err == nil {
			e.LabelDY = dy
		}
		if rp, err := finite(m["routing_padding"], fmt.Sprintf("arrows[%s].routing_padding", e.ID)); err == nil {
			e.RoutingPad = rp
		}
		if pc, err := finite(m["port_clearance"], fmt.Sprintf("arrows[%s].port_clearance", e.ID)); err == nil {
			e.PortClearance = pc
		}
		if ms, err := finite(m["motion_stage"], fmt.Sprintf("arrows[%s].motion_stage", e.ID)); err == nil {
			e.MotionStage = int(ms)
		}
		if mo, err := finite(m["motion_order"], fmt.Sprintf("arrows[%s].motion_order", e.ID)); err == nil {
			e.MotionOrder = int(mo)
		}

		// route_points
		if rps := asList(m["route_points"]); rps != nil {
			for ri, rp := range rps {
				pt, ok := rp.([]any)
				if !ok || len(pt) != 2 {
					return nil, diagramError("arrows[%s].route_points[%d] must be [x, y]", e.ID, ri)
				}
				x, err := finite(pt[0], fmt.Sprintf("arrows[%s].route_points[%d][0]", e.ID, ri))
				if err != nil {
					return nil, err
				}
				y, err := finite(pt[1], fmt.Sprintf("arrows[%s].route_points[%d][1]", e.ID, ri))
				if err != nil {
					return nil, err
				}
				e.RoutePoints = append(e.RoutePoints, [2]float64{x, y})
			}
		}
		// corridor hints
		if cxs := asList(m["corridor_x"]); cxs != nil {
			for _, cx := range cxs {
				if v, err := finite(cx, fmt.Sprintf("arrows[%s].corridor_x", e.ID)); err == nil {
					e.CorridorX = append(e.CorridorX, v)
				}
			}
		}
		if cys := asList(m["corridor_y"]); cys != nil {
			for _, cy := range cys {
				if v, err := finite(cy, fmt.Sprintf("arrows[%s].corridor_y", e.ID)); err == nil {
					e.CorridorY = append(e.CorridorY, v)
				}
			}
		}

		// validate source/target exist
		if e.Source != "" {
			if _, found := usedIDs[e.Source]; !found {
				return nil, diagramError("arrows[%s].source references unknown node: %s", e.ID, e.Source)
			}
		}
		if e.Target != "" {
			if _, found := usedIDs[e.Target]; !found {
				return nil, diagramError("arrows[%s].target references unknown node: %s", e.ID, e.Target)
			}
		}
		d.Edges = append(d.Edges, e)
	}

	// legend
	for _, raw := range asList(data["legend"]) {
		m := asMap(raw)
		if m == nil {
			continue
		}
		d.Legend = append(d.Legend, LegendEntry{
			Flow:  asString(m["flow"]),
			Label: asString(m["label"]),
		})
	}

	// sort nodes/edges by ID for determinism
	sort.Slice(d.Nodes, func(i, j int) bool { return d.Nodes[i].ID < d.Nodes[j].ID })
	sort.Slice(d.Edges, func(i, j int) bool { return d.Edges[i].ID < d.Edges[j].ID })

	return d, nil
}

// Bounds returns the Bounds of a node.
func (n Node) Bounds() (left, top, right, bottom float64) {
	return n.X, n.Y, n.X + n.Width, n.Y + n.Height
}
