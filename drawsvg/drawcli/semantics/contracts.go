// Package semantics provides style resolution and engineering semantic contracts.
// Ported from semantic_contracts.py (694 lines).
package semantics

import (
	"fmt"
	"math"
	"strings"
)

// StyleInfo describes a resolved visual style.
type StyleInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Profile     string `json:"profile"` // "generic", "c4-review", "cloud-fabric", "event-transit", "ops-pulse"
	VisualTheme string `json:"visual_theme"`
}

// ContractError is raised when a diagram violates a selected engineering contract.
type ContractError struct {
	Code    string
	Message string
}

func (e *ContractError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Message
}

func fail(code, format string, args ...any) *ContractError {
	return &ContractError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Style names index 1-12.
var StyleNames = map[int]string{
	1:  "Flat Icon",
	2:  "Dark Terminal",
	3:  "Blueprint",
	4:  "Notion Clean",
	5:  "Glassmorphism",
	6:  "Claude Official",
	7:  "OpenAI",
	8:  "Dark Luxury",
	9:  "C4 Review Canvas",
	10: "Cloud Fabric",
	11: "Event Transit",
	12: "Ops Pulse",
}

var styleDefaultProfiles = map[int]string{
	9:  "c4-review",
	10: "cloud-fabric",
	11: "event-transit",
	12: "ops-pulse",
}

var profileAliases = map[string]string{
	"generic": "generic", "none": "generic",
	"c4": "c4-review", "c4 review": "c4-review", "c4 review canvas": "c4-review",
	"architecture review board": "c4-review", "c4 评审画布": "c4-review", "架构评审画布": "c4-review", "c4-review": "c4-review",
	"cloud": "cloud-fabric", "cloud deployment": "cloud-fabric", "deployment topology": "cloud-fabric",
	"multi region deployment map": "cloud-fabric", "云部署拓扑": "cloud-fabric", "多区域部署图": "cloud-fabric",
	"cloud fabric": "cloud-fabric", "cloud-fabric": "cloud-fabric",
	"event stream": "event-transit", "event-stream": "event-transit", "event metro map": "event-transit",
	"topic rail map": "event-transit", "事件地铁图": "event-transit", "事件轨道图": "event-transit",
	"kafka": "event-transit", "event transit": "event-transit", "event-transit": "event-transit",
	"observability": "ops-pulse", "reliability pulse": "ops-pulse", "golden signals trace": "ops-pulse",
	"可靠性脉冲": "ops-pulse", "sre trace 评审": "ops-pulse", "otel": "ops-pulse",
	"ops pulse": "ops-pulse", "ops-pulse": "ops-pulse",
}

var styleAliases = func() map[string]int {
	m := make(map[string]int)
	for id, name := range StyleNames {
		tok := token(name)
		m[tok] = id
		m[fmt.Sprintf("style %d", id)] = id
		m[fmt.Sprintf("风格 %d", id)] = id
		m[fmt.Sprintf("风格%d", id)] = id
	}
	m["flat"] = 1
	m["terminal"] = 2
	m["dark terminal"] = 2
	m["notion"] = 4
	m["glass"] = 5
	m["claude"] = 6
	m["openai official"] = 7
	m["review canvas"] = 9
	m["c4 canvas"] = 9
	m["c4 review"] = 9
	m["adr review canvas"] = 9
	m["c4 评审"] = 9
	m["c4 评审画布"] = 9
	m["adr 评审图"] = 9
	m["架构评审画布"] = 9
	m["职责边界评审图"] = 9
	m["cloud deployment"] = 10
	m["deployment topology"] = 10
	m["multi region deployment map"] = 10
	m["云部署拓扑"] = 10
	m["多区域部署图"] = 10
	m["region vpc 归属图"] = 10
	m["event stream"] = 11
	m["event metro"] = 11
	m["event metro map"] = 11
	m["kafka topology"] = 11
	m["事件轨道图"] = 11
	m["事件地铁图"] = 11
	m["topic 线路图"] = 11
	m["kafka 拓扑图"] = 11
	m["sre"] = 12
	m["observability"] = 12
	m["reliability pulse"] = 12
	m["incident investigation view"] = 12
	m["sre trace review"] = 12
	m["golden signals trace"] = 12
	m["运维脉冲图"] = 12
	m["可靠性脉冲"] = 12
	m["事故排查视图"] = 12
	m["黄金信号追踪图"] = 12
	return m
}()

func token(s string) string {
	parts := strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(strings.ToLower(strings.TrimSpace(s))))
	return strings.Join(parts, " ")
}

// ResolveStyleIndex resolves numeric/name selectors and returns the style ID.
func ResolveStyleIndex(data map[string]any) (int, error) {
	type selector struct {
		field string
		value any
	}
	var selectors []selector

	if v, ok := data["style"]; ok {
		selectors = append(selectors, selector{"style", v})
	}
	if v, ok := data["visual_theme"]; ok {
		selectors = append(selectors, selector{"visual_theme", v})
	}
	if len(selectors) == 0 {
		return 1, nil
	}

	resolved := make(map[string]int)
	for _, sel := range selectors {
		var styleID int
		switch v := sel.value.(type) {
		case bool:
			return 0, fail("STYLE_SELECTOR", "%s must be a style id or name", sel.field)
		case int:
			styleID = v
		case float64:
			styleID = int(v)
		default:
			text := strings.TrimSpace(fmt.Sprint(v))
			if id, err := fmt.Sscanf(text, "%d", &styleID); err == nil && id == 1 {
				// numeric string
			} else {
				normalized := token(text)
				var found bool
				styleID, found = styleAliases[normalized]
				if !found {
					return 0, fail("STYLE_SELECTOR", "unsupported %s: %v", sel.field, v)
				}
			}
		}
		if _, ok := StyleNames[styleID]; !ok {
			return 0, fail("STYLE_SELECTOR", "unsupported %s: %v", sel.field, sel.value)
		}
		resolved[sel.field] = styleID
	}

	// check conflicts
	var firstID int
	for _, id := range resolved {
		if firstID == 0 {
			firstID = id
		} else if id != firstID {
			return 0, fail("STYLE_SELECTOR_CONFLICT", "conflicting style selectors: %v", resolved)
		}
	}
	return firstID, nil
}

// SemanticReport is returned by ValidateSemanticContract.
type SemanticReport struct {
	Ok          bool           `json:"ok"`
	Style       int            `json:"style"`
	VisualTheme string         `json:"visual_theme"`
	Profile     string         `json:"profile"`
	Details     map[string]any `json:"details"`
}

// ValidateSemanticContract validates and enriches a normalized diagram payload.
func ValidateSemanticContract(styleIndex int, profileHint string, data map[string]any) (*SemanticReport, error) {
	profile := profileHint
	if profile == "" || profile == "auto" {
		if p, ok := styleDefaultProfiles[styleIndex]; ok {
			profile = p
		} else {
			profile = "generic"
		}
	} else {
		normalized := token(profile)
		if a, ok := profileAliases[normalized]; ok {
			profile = a
		} else {
			return nil, fail("SEMANTIC_PROFILE", "unsupported semantic_profile: %s", profile)
		}
	}
	data["semantic_profile"] = profile

	report := &SemanticReport{
		Style:       styleIndex,
		VisualTheme: StyleNames[styleIndex],
		Profile:     profile,
		Details:     make(map[string]any),
	}

	var contractErr error
	switch profile {
	case "c4-review":
		contractErr = validateC4(data)
	case "cloud-fabric":
		contractErr = validateCloud(data)
	case "event-transit":
		contractErr = validateEvent(data)
	case "ops-pulse":
		contractErr = validateOps(data)
	}
	if contractErr != nil {
		return nil, contractErr
	}

	report.Ok = true
	return report, nil
}

// helper: require a non-empty text field
func requireText(m map[string]any, field, path string) (string, error) {
	v, ok := m[field]
	if !ok || v == nil {
		return "", fail("SEMANTIC_REQUIRED", "%s.%s is required", path, field)
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" {
		return "", fail("SEMANTIC_REQUIRED", "%s.%s is required", path, field)
	}
	return s, nil
}

func requireInt(m map[string]any, field, path string) (int, error) {
	v, ok := m[field]
	if !ok {
		return 0, fail("SEMANTIC_REQUIRED", "%s.%s is required", path, field)
	}
	switch n := v.(type) {
	case int:
		return n, nil
	case float64:
		return int(n), nil
	default:
		return 0, fail("SEMANTIC_NUMBER", "%s.%s must be a number", path, field)
	}
}

func requireInteger(m map[string]any, field, path string) (int, error) {
	v, ok := m[field]
	if !ok || v == nil {
		return 0, fail("SEMANTIC_REQUIRED", "%s.%s is required", path, field)
	}
	switch n := v.(type) {
	case int:
		return n, nil
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n {
			return 0, fail("EVENT_STATION_ORDER", "%s.%s must be an integer", path, field)
		}
		return int(n), nil
	default:
		return 0, fail("EVENT_STATION_ORDER", "%s.%s must be an integer", path, field)
	}
}

func setDefault(m map[string]any, key string, value any) {
	if _, exists := m[key]; !exists {
		m[key] = value
	}
}

func requireGraphEndpoints(data map[string]any, nodes map[string]map[string]any) (map[string]map[string]any, error) {
	edges := getEdgeMap(data)
	for id, edge := range edges {
		source, err := requireText(edge, "source", fmt.Sprintf("arrows[%s]", id))
		if err != nil {
			return nil, err
		}
		target, err := requireText(edge, "target", fmt.Sprintf("arrows[%s]", id))
		if err != nil {
			return nil, err
		}
		if _, ok := nodes[source]; !ok {
			return nil, fail("SEMANTIC_EDGE_ENDPOINT", "arrows[%s] references unknown source %s", id, source)
		}
		if _, ok := nodes[target]; !ok {
			return nil, fail("SEMANTIC_EDGE_ENDPOINT", "arrows[%s] references unknown target %s", id, target)
		}
	}
	return edges, nil
}

func boundsFromMap(m map[string]any, path string) (b [4]float64, err error) {
	x, _ := toFloat64(m["x"])
	y, _ := toFloat64(m["y"])
	w, _ := toFloat64(m["width"])
	h, _ := toFloat64(m["height"])
	if w <= 0 || h <= 0 {
		return b, fail("SEMANTIC_BOUNDS", "%s must have positive width and height", path)
	}
	return [4]float64{x, y, x + w, y + h}, nil
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		if math.IsInf(n, 0) || math.IsNaN(n) {
			return 0, false
		}
		return n, true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}

func inside(inner [4]float64, outer [4]float64, inset float64) bool {
	return inner[0] >= outer[0]+inset &&
		inner[1] >= outer[1]+inset &&
		inner[2] <= outer[2]-inset &&
		inner[3] <= outer[3]-inset
}

func rectGap(first, second [4]float64) float64 {
	h := math.Max(math.Max(first[0]-second[2], second[0]-first[2]), 0)
	v := math.Max(math.Max(first[1]-second[3], second[1]-first[3]), 0)
	return math.Hypot(h, v)
}

// --- C4 validation ---

func validateC4(data map[string]any) error {
	if getString(data, "diagram_type") != "c4" {
		return fail("C4_DIAGRAM_TYPE", "diagram_type must be 'c4'")
	}
	level, err := requireText(data, "c4_level", "diagram")
	if err != nil {
		return err
	}
	if level != "context" && level != "container" && level != "component" {
		return fail("C4_LEVEL", "unsupported c4_level: %s", level)
	}
	if _, err := requireText(data, "title", "diagram"); err != nil {
		return err
	}
	if _, err := requireText(data, "scope", "diagram"); err != nil {
		return err
	}
	seed, err := requireInt(data, "rough_seed", "diagram")
	if err != nil {
		return err
	}
	legend, _ := data["legend"].([]any)
	if len(legend) == 0 {
		return fail("C4_LEGEND", "a non-empty legend is required")
	}

	allowed := map[string]map[string]bool{
		"context":   {"person": true, "software_system": true, "external_system": true},
		"container": {"person": true, "software_system": true, "external_system": true, "container": true},
		"component": {"person": true, "software_system": true, "external_system": true, "container": true, "component": true},
	}[level]

	// build container boundaries
	boundaries := make(map[string]map[string]any)
	for _, raw := range getList(data, "containers") {
		if m, ok := raw.(map[string]any); ok {
			boundaries[fmt.Sprint(m["id"])] = m
		}
	}

	nodes := getNodeMap(data)
	for id, node := range nodes {
		c4Type, err := requireText(node, "c4_type", fmt.Sprintf("nodes[%s]", id))
		if err != nil {
			return err
		}
		if !allowed[c4Type] {
			return fail("C4_MIXED_ABSTRACTION", "%s type %q is invalid for %s view", id, c4Type, level)
		}
		if _, err := requireText(node, "label", fmt.Sprintf("nodes[%s]", id)); err != nil {
			return err
		}
		if _, err := requireText(node, "description", fmt.Sprintf("nodes[%s]", id)); err != nil {
			return err
		}
		if c4Type == "container" || c4Type == "component" {
			if _, err := requireText(node, "technology", fmt.Sprintf("nodes[%s]", id)); err != nil {
				return err
			}
			var b [4]float64
			b, err = boundsFromMap(node, fmt.Sprintf("nodes[%s]", id))
			if err != nil {
				return err
			}
			if b[2]-b[0] < 170 || b[3]-b[1] < 96 {
				return fail("C4_CARD_SIZE", "%s must be at least 170x96", id)
			}
		}
		parent := getString(node, "parent")
		if parent != "" {
			pBoundary, ok := boundaries[parent]
			if !ok {
				return fail("C4_PARENT", "%s references unknown boundary %s", id, parent)
			}
			var nb, pb [4]float64
			nb, err = boundsFromMap(node, fmt.Sprintf("nodes[%s]", id))
			pb, err = boundsFromMap(pBoundary, fmt.Sprintf("containers[%s]", parent))
			if err == nil && !inside(nb, pb, 20) {
				return fail("C4_BOUNDARY_ESCAPE", "%s must stay 20px inside %s", id, parent)
			}
		}
	}

	edges := getEdgeMap(data)
	for _, edge := range edges {
		if _, err := requireText(edge, "label", fmt.Sprintf("arrows[%s]", edge["id"])); err != nil {
			return err
		}
		if _, err := requireText(edge, "protocol", fmt.Sprintf("arrows[%s]", edge["id"])); err != nil {
			return err
		}
	}

	data["c4_valid"] = map[string]any{
		"level":    level,
		"seed":     seed,
		"elements": len(nodes),
	}
	return nil
}

const cloudManifestVersion = "2026.07-neutral.1"

type cloudIcon struct{ badge, color, glyph string }

var cloudIcons = map[string]cloudIcon{
	"generic:traffic":       {"EDGE", "#2563eb", "globe"},
	"generic:gateway":       {"GW", "#7c3aed", "gateway"},
	"generic:compute":       {"APP", "#059669", "compute"},
	"generic:database":      {"DB", "#ea580c", "database"},
	"generic:stream":        {"BUS", "#db2777", "stream"},
	"generic:observability": {"OBS", "#0891b2", "observe"},
}

// --- Cloud validation ---
func validateCloud(data map[string]any) error {
	if getString(data, "diagram_type") != "deployment" {
		return fail("CLOUD_DIAGRAM_TYPE", "diagram_type must be 'deployment'")
	}
	platform, err := requireText(data, "platform_profile", "diagram")
	if err != nil {
		return err
	}
	if !map[string]bool{"provider-neutral": true, "aws": true, "azure": true, "gcp": true, "kubernetes": true}[platform] {
		return fail("CLOUD_PLATFORM", "unsupported platform_profile: %s", platform)
	}
	if getString(data, "icon_manifest_version") != cloudManifestVersion {
		return fail("CLOUD_ICON_VERSION", "icon_manifest_version must be %s", cloudManifestVersion)
	}

	containers := make(map[string]map[string]any)
	for _, raw := range getList(data, "containers") {
		container, ok := raw.(map[string]any)
		if !ok {
			return fail("CLOUD_BOUNDARY", "container must be an object")
		}
		id, err := requireText(container, "id", "containers")
		if err != nil {
			return err
		}
		if _, exists := containers[id]; exists {
			return fail("CLOUD_BOUNDARY", "duplicate container id: %s", id)
		}
		containers[id] = container
	}
	if len(containers) == 0 {
		return fail("CLOUD_BOUNDARY", "at least one region deployment boundary is required")
	}
	regionFound := false
	for id, container := range containers {
		if getString(container, "deployment_kind") == "region" {
			regionFound = true
		}
		if _, err := requireText(container, "deployment_kind", fmt.Sprintf("containers[%s]", id)); err != nil {
			return err
		}
		if _, err := boundsFromMap(container, fmt.Sprintf("containers[%s]", id)); err != nil {
			return err
		}
	}
	if !regionFound {
		return fail("CLOUD_BOUNDARY", "at least one region deployment boundary is required")
	}

	depths := make(map[string]int, len(containers))
	var depth func(string, map[string]bool) (int, error)
	depth = func(id string, trail map[string]bool) (int, error) {
		if cached, ok := depths[id]; ok {
			return cached, nil
		}
		if trail[id] {
			return 0, fail("CLOUD_BOUNDARY_CYCLE", "container cycle includes %s", id)
		}
		trail[id] = true
		parent := getString(containers[id], "parent")
		result := 1
		if parent != "" {
			if _, ok := containers[parent]; !ok {
				return 0, fail("CLOUD_BOUNDARY_PARENT", "%s references unknown parent %s", id, parent)
			}
			parentDepth, err := depth(parent, trail)
			if err != nil {
				return 0, err
			}
			result = parentDepth + 1
			childBounds, _ := boundsFromMap(containers[id], fmt.Sprintf("containers[%s]", id))
			parentBounds, _ := boundsFromMap(containers[parent], fmt.Sprintf("containers[%s]", parent))
			if !inside(childBounds, parentBounds, 16) {
				return 0, fail("CLOUD_BOUNDARY_ESCAPE", "%s must be inset inside %s", id, parent)
			}
		}
		delete(trail, id)
		depths[id] = result
		return result, nil
	}
	maxDepth := 0
	for id := range containers {
		value, err := depth(id, make(map[string]bool))
		if err != nil {
			return err
		}
		maxDepth = max(maxDepth, value)
	}
	if maxDepth > 4 {
		return fail("CLOUD_BOUNDARY_DEPTH", "deployment nesting depth cannot exceed 4")
	}
	for firstID, first := range containers {
		for secondID, second := range containers {
			if firstID >= secondID || getString(first, "parent") != getString(second, "parent") {
				continue
			}
			firstBounds, _ := boundsFromMap(first, fmt.Sprintf("containers[%s]", firstID))
			secondBounds, _ := boundsFromMap(second, fmt.Sprintf("containers[%s]", secondID))
			if rectGap(firstBounds, secondBounds) < 16 {
				return fail("CLOUD_BOUNDARY_GAP", "siblings %s and %s need 16px clearance", firstID, secondID)
			}
		}
	}

	nodes := getNodeMap(data)
	for id, node := range nodes {
		deploymentID, err := requireText(node, "deployment_id", fmt.Sprintf("nodes[%s]", id))
		if err != nil {
			return err
		}
		container, ok := containers[deploymentID]
		if !ok {
			return fail("CLOUD_DEPLOYMENT", "%s references unknown deployment %s", id, deploymentID)
		}
		nodeBounds, err := boundsFromMap(node, fmt.Sprintf("nodes[%s]", id))
		if err != nil {
			return err
		}
		containerBounds, _ := boundsFromMap(container, fmt.Sprintf("containers[%s]", deploymentID))
		if !inside(nodeBounds, containerBounds, 20) {
			return fail("CLOUD_NODE_ESCAPE", "%s must stay 20px inside %s", id, deploymentID)
		}
		iconID, err := requireText(node, "icon_id", fmt.Sprintf("nodes[%s]", id))
		if err != nil {
			return err
		}
		icon, ok := cloudIcons[iconID]
		if !ok {
			return fail("CLOUD_ICON_UNKNOWN", "unknown icon_id: %s", iconID)
		}
		setDefault(node, "icon_badge", icon.badge)
		setDefault(node, "icon_color", icon.color)
		setDefault(node, "glyph", icon.glyph)
		setDefault(node, "icon_source", "builtin-neutral")
		setDefault(node, "icon_version", cloudManifestVersion)
		var labels []string
		for current := deploymentID; current != ""; current = getString(containers[current], "parent") {
			labels = append([]string{getString(containers[current], "label")}, labels...)
		}
		setDefault(node, "deployment_path", strings.Join(labels, " › "))
	}
	edges, err := requireGraphEndpoints(data, nodes)
	if err != nil {
		return err
	}
	for id, edge := range edges {
		source := nodes[getString(edge, "source")]
		target := nodes[getString(edge, "target")]
		if getString(source, "deployment_id") != getString(target, "deployment_id") {
			if _, err := requireText(edge, "via", fmt.Sprintf("arrows[%s]", id)); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- Event validation ---
func validateEvent(data map[string]any) error {
	if getString(data, "diagram_type") != "event_stream" {
		return fail("EVENT_DIAGRAM_TYPE", "diagram_type must be 'event_stream'")
	}
	topics := getList(data, "topics")
	if len(topics) == 0 {
		return fail("EVENT_TOPICS", "topics must be a non-empty array")
	}
	if len(topics) > 4 {
		return fail("EVENT_TOPIC_LIMIT", "showcase diagrams support at most four topic rails")
	}
	topicColors := make(map[string]string, len(topics))
	for index, raw := range topics {
		topic, ok := raw.(map[string]any)
		if !ok {
			return fail("EVENT_TOPIC", "topics[%d] must be an object", index)
		}
		id, err := requireText(topic, "id", fmt.Sprintf("topics[%d]", index))
		if err != nil {
			return err
		}
		color, err := requireText(topic, "color", fmt.Sprintf("topics[%d]", index))
		if err != nil {
			return err
		}
		if _, duplicate := topicColors[id]; duplicate {
			return fail("EVENT_TOPIC_DUPLICATE", "duplicate topic id: %s", id)
		}
		topicColors[id] = color
	}
	nodes := getNodeMap(data)
	allowedRoles := map[string]bool{"producer": true, "station": true, "junction": true, "consumer": true, "dlq": true, "state_store": true}
	orders := make(map[string]string)
	for id, node := range nodes {
		role, err := requireText(node, "transit_role", fmt.Sprintf("nodes[%s]", id))
		if err != nil {
			return err
		}
		if !allowedRoles[role] {
			return fail("EVENT_ROLE", "unsupported transit_role %q on %s", role, id)
		}
		topicID := getString(node, "topic_id")
		if topicID != "" {
			color, exists := topicColors[topicID]
			if !exists {
				return fail("EVENT_TOPIC_UNKNOWN", "%s references unknown topic %s", id, topicID)
			}
			setDefault(node, "rail_color", color)
		}
		if role == "station" || role == "junction" {
			if _, err := requireText(node, "operation", fmt.Sprintf("nodes[%s]", id)); err != nil {
				return err
			}
		}
		if role == "consumer" {
			if _, err := requireText(node, "consumer_group", fmt.Sprintf("nodes[%s]", id)); err != nil {
				return err
			}
		}
		if role != "dlq" && role != "state_store" {
			order, err := requireInteger(node, "station_order", fmt.Sprintf("nodes[%s]", id))
			if err != nil {
				return err
			}
			key := fmt.Sprintf("%s:%d", topicID, order)
			if previous, exists := orders[key]; exists {
				return fail("EVENT_STATION_ORDER", "duplicate order %d for topic %s (%s, %s)", order, topicID, previous, id)
			}
			orders[key] = id
		}
	}
	edges, err := requireGraphEndpoints(data, nodes)
	if err != nil {
		return err
	}
	railOutgoing := make(map[string]int)
	for id, edge := range edges {
		transitType, err := requireText(edge, "transit_type", fmt.Sprintf("arrows[%s]", id))
		if err != nil {
			return err
		}
		source := nodes[getString(edge, "source")]
		target := nodes[getString(edge, "target")]
		switch transitType {
		case "rail":
			topicID, err := requireText(edge, "topic_id", fmt.Sprintf("arrows[%s]", id))
			if err != nil {
				return err
			}
			if getString(source, "topic_id") != topicID || getString(target, "topic_id") != topicID {
				return fail("EVENT_TOPIC_DRIFT", "%s must connect nodes on topic %s", id, topicID)
			}
			sourceOrder, _ := requireInteger(source, "station_order", fmt.Sprintf("nodes[%s]", getString(source, "id")))
			targetOrder, _ := requireInteger(target, "station_order", fmt.Sprintf("nodes[%s]", getString(target, "id")))
			if targetOrder != sourceOrder+1 {
				return fail("EVENT_RAIL_ORDER", "%s must connect adjacent increasing stations", id)
			}
			sourceBounds, _ := boundsFromMap(source, "source")
			targetBounds, _ := boundsFromMap(target, "target")
			if math.Abs((sourceBounds[1]+sourceBounds[3])-(targetBounds[1]+targetBounds[3])) > 1e-6 {
				return fail("EVENT_RAIL_ALIGNMENT", "%s rail endpoints must share one horizontal centerline", id)
			}
			if targetBounds[0]-sourceBounds[2] < 64 {
				return fail("EVENT_RAIL_LENGTH", "%s rail must have at least 64px clearance", id)
			}
			if getString(edge, "source_port") != "right" || getString(edge, "target_port") != "left" {
				return fail("EVENT_RAIL_PORT", "%s must use right-to-left ports", id)
			}
			railOutgoing[getString(edge, "source")]++
		case "dead_letter":
			if getString(target, "transit_role") != "dlq" {
				return fail("EVENT_DLQ_TARGET", "%s must target a dlq node", id)
			}
			setDefault(edge, "dashed", true)
		case "branch":
			if getString(source, "transit_role") != "junction" {
				return fail("EVENT_BRANCH_JUNCTION", "%s must depart from a junction node", id)
			}
		case "publish", "consume", "retry", "state":
		default:
			return fail("EVENT_EDGE_TYPE", "unsupported transit_type: %s", transitType)
		}
	}
	for id, count := range railOutgoing {
		if count > 1 && getString(nodes[id], "transit_role") != "junction" {
			return fail("EVENT_BRANCH_JUNCTION", "%s branches without a junction role", id)
		}
	}
	return nil
}

// --- Ops validation ---
func validateOps(data map[string]any) error {
	if getString(data, "diagram_type") != "observability" {
		return fail("OPS_DIAGRAM_TYPE", "diagram_type must be 'observability'")
	}
	observationWindow, err := requireText(data, "observation_window", "diagram")
	if err != nil {
		return err
	}
	nodes := getNodeMap(data)
	edges, err := requireGraphEndpoints(data, nodes)
	if err != nil {
		return err
	}
	services := make(map[string]map[string]any)
	for id, node := range nodes {
		if getString(node, "ops_role") != "service" {
			continue
		}
		services[id] = node
		status := getString(node, "status")
		if _, ok := map[string]bool{"ok": true, "warn": true, "critical": true, "unknown": true}[status]; !ok {
			return fail("OPS_SERVICE_STATUS", "unsupported status %q on %s", status, id)
		}
		if _, err := requireText(node, "status_label", fmt.Sprintf("nodes[%s]", id)); err != nil {
			return err
		}
		cardBounds, err := boundsFromMap(node, fmt.Sprintf("nodes[%s]", id))
		if err != nil {
			return err
		}
		if cardBounds[2]-cardBounds[0] < 180 || cardBounds[3]-cardBounds[1] < 108 {
			return fail("OPS_CARD_SIZE", "%s must be at least 180x108", id)
		}
		signals, ok := node["signals"].(map[string]any)
		if !ok || len(signals) != 4 {
			return fail("OPS_GOLDEN_SIGNALS", "%s must define latency, traffic, errors, and saturation", id)
		}
		badges := make([]any, 0, 4)
		for _, name := range []string{"latency", "traffic", "errors", "saturation"} {
			rawSignal, exists := signals[name]
			signal, ok := rawSignal.(map[string]any)
			if !exists || !ok {
				return fail("OPS_SIGNAL", "%s.%s must be an object", id, name)
			}
			value, err := requireText(signal, "value", fmt.Sprintf("nodes[%s].signals.%s", id, name))
			if err != nil {
				return err
			}
			unit, err := requireText(signal, "unit", fmt.Sprintf("nodes[%s].signals.%s", id, name))
			if err != nil {
				return err
			}
			window, err := requireText(signal, "window", fmt.Sprintf("nodes[%s].signals.%s", id, name))
			if err != nil {
				return err
			}
			if window != observationWindow {
				return fail("OPS_OBSERVATION_WINDOW", "%s.%s window %q must match diagram observation_window %q", id, name, window, observationWindow)
			}
			metricStatus := getString(signal, "status")
			if _, ok := map[string]bool{"ok": true, "warn": true, "critical": true, "unknown": true}[metricStatus]; !ok {
				return fail("OPS_SIGNAL_STATUS", "unsupported status %q on %s.%s", metricStatus, id, name)
			}
			badges = append(badges, map[string]any{"name": name, "value": value, "unit": unit, "window": window, "status": metricStatus})
		}
		node["metric_badges"] = badges
	}
	if len(services) == 0 || len(services) > 12 {
		return fail("OPS_SERVICE_LIMIT", "an Ops Pulse view requires 1-12 service nodes")
	}

	criticalPath, ok := data["critical_path"].([]any)
	if !ok || len(criticalPath) == 0 {
		return fail("OPS_CRITICAL_PATH", "critical_path must be a non-empty ordered edge-id list")
	}
	previousTarget := ""
	visitedServices := make(map[string]bool)
	for index, rawID := range criticalPath {
		edgeID := strings.TrimSpace(fmt.Sprint(rawID))
		edge, ok := edges[edgeID]
		if !ok {
			return fail("OPS_CRITICAL_PATH", "unknown critical edge: %s", edgeID)
		}
		if getString(edge, "edge_kind") != "" && getString(edge, "edge_kind") != "business" {
			return fail("OPS_CRITICAL_PATH", "%s must be a business edge", edgeID)
		}
		if index > 0 && getString(edge, "source") != previousTarget {
			return fail("OPS_CRITICAL_PATH", "%s does not continue the critical path", edgeID)
		}
		if _, ok := services[getString(edge, "source")]; !ok {
			return fail("OPS_CRITICAL_PATH", "%s source must be a service", edgeID)
		}
		if _, ok := services[getString(edge, "target")]; !ok {
			return fail("OPS_CRITICAL_PATH", "%s target must be a service", edgeID)
		}
		if index == 0 {
			visitedServices[getString(edge, "source")] = true
		}
		if visitedServices[getString(edge, "target")] {
			return fail("OPS_CRITICAL_PATH", "critical path repeats service %s", getString(edge, "target"))
		}
		visitedServices[getString(edge, "target")] = true
		edge["critical"] = true
		pathID := getString(data, "critical_path_id")
		if pathID == "" {
			pathID = "critical-1"
		}
		edge["critical_path_id"] = pathID
		edge["critical_hop"] = index + 1
		edge["critical_hops"] = len(criticalPath)
		previousTarget = getString(edge, "target")
	}
	businessFlows := make(map[string]bool)
	telemetryFlows := make(map[string]bool)
	for _, edge := range edges {
		if getString(edge, "edge_kind") == "business" {
			businessFlows[getString(edge, "flow")] = true
		}
		if getString(edge, "edge_kind") == "telemetry" {
			telemetryFlows[getString(edge, "flow")] = true
		}
	}
	for flow := range businessFlows {
		if telemetryFlows[flow] {
			return fail("OPS_FLOW_SEMANTICS", "business and telemetry edges must use different flow tokens")
		}
	}
	if err := validateTraceSpans(nodes); err != nil {
		return err
	}
	return nil
}

func validateTraceSpans(nodes map[string]map[string]any) error {
	spans := make(map[string]map[string]any)
	for nodeID, node := range nodes {
		if getString(node, "ops_role") != "trace_span" {
			continue
		}
		spanID, err := requireText(node, "span_id", fmt.Sprintf("nodes[%s]", nodeID))
		if err != nil {
			return err
		}
		if _, exists := spans[spanID]; exists {
			return fail("OPS_SPAN_ID", "trace span ids must be unique")
		}
		spans[spanID] = node
	}
	if len(spans) == 0 {
		return fail("OPS_TRACE_REQUIRED", "an Ops Pulse view requires one correlated trace waterfall")
	}
	var root map[string]any
	roots := 0
	for spanID, span := range spans {
		start, ok := toFloat64(span["start_ms"])
		if !ok {
			return fail("SEMANTIC_NUMBER", "spans[%s].start_ms must be finite", spanID)
		}
		duration, ok := toFloat64(span["duration_ms"])
		if !ok || duration <= 0 {
			return fail("OPS_SPAN_DURATION", "%s duration must be positive", spanID)
		}
		parentID := getString(span, "parent_span")
		if parentID == "" {
			roots++
			root = span
			continue
		}
		parent, exists := spans[parentID]
		if !exists {
			return fail("OPS_SPAN_PARENT", "%s references unknown parent span %s", spanID, parentID)
		}
		parentStart, _ := toFloat64(parent["start_ms"])
		parentDuration, _ := toFloat64(parent["duration_ms"])
		if start < parentStart || start+duration > parentStart+parentDuration {
			return fail("OPS_SPAN_COVERAGE", "%s must be contained by %s", spanID, parentID)
		}
	}
	if roots != 1 {
		return fail("OPS_SPAN_ROOT", "trace waterfall must contain exactly one root span")
	}
	for spanID, span := range spans {
		seen := make(map[string]bool)
		currentID := spanID
		current := span
		for getString(current, "parent_span") != "" {
			if seen[currentID] {
				return fail("OPS_SPAN_CYCLE", "trace parent cycle contains %s", currentID)
			}
			seen[currentID] = true
			currentID = getString(current, "parent_span")
			current = spans[currentID]
		}
	}
	rootBounds, err := boundsFromMap(root, "root_span")
	if err != nil {
		return err
	}
	rootStart, _ := toFloat64(root["start_ms"])
	rootDuration, _ := toFloat64(root["duration_ms"])
	pixelsPerMS := (rootBounds[2] - rootBounds[0]) / rootDuration
	originX := rootBounds[0] - rootStart*pixelsPerMS
	for spanID, span := range spans {
		bounds, err := boundsFromMap(span, fmt.Sprintf("spans[%s]", spanID))
		if err != nil {
			return err
		}
		start, _ := toFloat64(span["start_ms"])
		duration, _ := toFloat64(span["duration_ms"])
		if math.Abs(bounds[0]-(originX+start*pixelsPerMS)) > 1.5 || math.Abs((bounds[2]-bounds[0])-duration*pixelsPerMS) > 1.5 {
			return fail("OPS_SPAN_SCALE", "%s x/width must encode start_ms/duration_ms on the root time scale", spanID)
		}
	}
	return nil
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func getList(data map[string]any, key string) []any {
	v, _ := data[key].([]any)
	return v
}

func getNodeMap(data map[string]any) map[string]map[string]any {
	result := make(map[string]map[string]any)
	for _, raw := range getList(data, "nodes") {
		if m, ok := raw.(map[string]any); ok {
			id := fmt.Sprint(m["id"])
			result[id] = m
		}
	}
	return result
}

func getEdgeMap(data map[string]any) map[string]map[string]any {
	result := make(map[string]map[string]any)
	raw := asListAny(data, "arrows")
	if len(raw) == 0 {
		raw = asListAny(data, "edges")
	}
	for _, rawEdge := range raw {
		if m, ok := rawEdge.(map[string]any); ok {
			id := fmt.Sprint(m["id"])
			m["id"] = id
			result[id] = m
		}
	}
	return result
}

func asListAny(data map[string]any, key string) []any {
	v, _ := data[key].([]any)
	return v
}
