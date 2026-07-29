package validate

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"drawcli/geometry"
)

type svgShape struct {
	id     string
	bounds geometry.Bounds
}
type svgPath struct {
	id    string
	route []geometry.Point
}

type svgMatrix struct{ a, b, c, d, e, f float64 }

var identityMatrix = svgMatrix{a: 1, d: 1}
var transformPattern = regexp.MustCompile(`([A-Za-z]+)\s*\(([^)]*)\)`)
var numberPattern = regexp.MustCompile(`[-+]?(?:\d*\.\d+|\d+\.?)(?:[eE][-+]?\d+)?`)

func multiplyMatrix(left, right svgMatrix) svgMatrix {
	return svgMatrix{
		a: left.a*right.a + left.c*right.b,
		b: left.b*right.a + left.d*right.b,
		c: left.a*right.c + left.c*right.d,
		d: left.b*right.c + left.d*right.d,
		e: left.a*right.e + left.c*right.f + left.e,
		f: left.b*right.e + left.d*right.f + left.f,
	}
}

func transformSVGPoint(matrix svgMatrix, point geometry.Point) geometry.Point {
	return geometry.Point{X: matrix.a*point.X + matrix.c*point.Y + matrix.e, Y: matrix.b*point.X + matrix.d*point.Y + matrix.f}
}

func parseTransform(value string) svgMatrix {
	result := identityMatrix
	for _, match := range transformPattern.FindAllStringSubmatch(value, -1) {
		var values []float64
		for _, raw := range numberPattern.FindAllString(match[2], -1) {
			if number, err := strconv.ParseFloat(raw, 64); err == nil {
				values = append(values, number)
			}
		}
		current := identityMatrix
		switch strings.ToLower(match[1]) {
		case "matrix":
			if len(values) == 6 {
				current = svgMatrix{values[0], values[1], values[2], values[3], values[4], values[5]}
			}
		case "translate":
			if len(values) > 0 {
				current.e = values[0]
				if len(values) > 1 {
					current.f = values[1]
				}
			}
		case "scale":
			if len(values) > 0 {
				current.a, current.d = values[0], values[0]
				if len(values) > 1 {
					current.d = values[1]
				}
			}
		case "rotate":
			if len(values) > 0 {
				angle := values[0] * math.Pi / 180
				rotation := svgMatrix{a: math.Cos(angle), b: math.Sin(angle), c: -math.Sin(angle), d: math.Cos(angle)}
				if len(values) >= 3 {
					current = multiplyMatrix(multiplyMatrix(svgMatrix{a: 1, d: 1, e: values[1], f: values[2]}, rotation), svgMatrix{a: 1, d: 1, e: -values[1], f: -values[2]})
				} else {
					current = rotation
				}
			}
		case "skewx":
			if len(values) == 1 {
				current.c = math.Tan(values[0] * math.Pi / 180)
			}
		case "skewy":
			if len(values) == 1 {
				current.b = math.Tan(values[0] * math.Pi / 180)
			}
		}
		result = multiplyMatrix(result, current)
	}
	return result
}

func svgGeometry(content string) ([]svgShape, []svgPath, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	var shapes []svgShape
	var paths []svgPath
	type context struct {
		matrix svgMatrix
		inDefs bool
		role   string
	}
	stack := []context{{matrix: identityMatrix}}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return shapes, paths, nil
		}
		if err != nil {
			return nil, nil, err
		}
		switch element := token.(type) {
		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
			continue
		case xml.StartElement:
			start := element
			attrs := map[string]string{}
			for _, attr := range start.Attr {
				attrs[attr.Name.Local] = attr.Value
			}
			parent := stack[len(stack)-1]
			role := strings.ToLower(strings.TrimSpace(attrs["data-graph-role"]))
			if role == "" && inheritedRole(parent.role) {
				role = parent.role
			}
			current := context{matrix: multiplyMatrix(parent.matrix, parseTransform(attrs["transform"])), inDefs: parent.inDefs || start.Name.Local == "defs", role: role}
			stack = append(stack, current)
			if current.inDefs {
				continue
			}
			switch start.Name.Local {
			case "rect":
				if current.role != "node" {
					continue
				}
				x, okX := parseSVGNumber(attrs["x"])
				y, okY := parseSVGNumber(attrs["y"])
				w, okW := parseSVGNumber(attrs["width"])
				h, okH := parseSVGNumber(attrs["height"])
				if okX && okY && okW && okH && w > 0 && h > 0 {
					points := []geometry.Point{{X: x, Y: y}, {X: x + w, Y: y}, {X: x, Y: y + h}, {X: x + w, Y: y + h}}
					left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
					for _, point := range points {
						point = transformSVGPoint(current.matrix, point)
						left, top = math.Min(left, point.X), math.Min(top, point.Y)
						right, bottom = math.Max(right, point.X), math.Max(bottom, point.Y)
					}
					shapes = append(shapes, svgShape{attrs["id"], geometry.Bounds{Left: left, Top: top, Right: right, Bottom: bottom}})
				}
			case "circle", "ellipse":
				if current.role != "node" {
					continue
				}
				cx, okX := parseSVGNumber(attrs["cx"])
				cy, okY := parseSVGNumber(attrs["cy"])
				rx, okRX := parseSVGNumber(attrs["rx"])
				ry, okRY := parseSVGNumber(attrs["ry"])
				if start.Name.Local == "circle" {
					r, ok := parseSVGNumber(attrs["r"])
					rx, ry, okRX, okRY = r, r, ok, ok
				}
				if okX && okY && okRX && okRY && rx > 0 && ry > 0 {
					shapes = append(shapes, svgShape{attrs["id"], transformedBounds(current.matrix, []geometry.Point{{X: cx - rx, Y: cy - ry}, {X: cx + rx, Y: cy - ry}, {X: cx - rx, Y: cy + ry}, {X: cx + rx, Y: cy + ry}})})
				}
			case "polygon", "polyline":
				if current.role != "node" {
					continue
				}
				points := parseSVGPoints(attrs["points"])
				if len(points) >= 2 {
					shapes = append(shapes, svgShape{attrs["id"], transformedBounds(current.matrix, points)})
				}
			case "path":
				if current.role != "edge" && current.role != "node" {
					continue
				}
				route, err := parseSVGPath(attrs["d"])
				if err != nil {
					return nil, nil, fmt.Errorf("path %q: %w", attrs["id"], err)
				}
				if len(route) > 1 {
					for index := range route {
						route[index] = transformSVGPoint(current.matrix, route[index])
					}
					if current.role == "edge" {
						paths = append(paths, svgPath{attrs["id"], route})
					} else {
						shapes = append(shapes, svgShape{attrs["id"], transformedBounds(identityMatrix, route)})
					}
				}
			}
		}
	}
}

func transformedBounds(matrix svgMatrix, points []geometry.Point) geometry.Bounds {
	left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, point := range points {
		point = transformSVGPoint(matrix, point)
		left, top = math.Min(left, point.X), math.Min(top, point.Y)
		right, bottom = math.Max(right, point.X), math.Max(bottom, point.Y)
	}
	return geometry.Bounds{Left: left, Top: top, Right: right, Bottom: bottom}
}

func parseSVGPoints(value string) []geometry.Point {
	values := numberPattern.FindAllString(value, -1)
	points := make([]geometry.Point, 0, len(values)/2)
	for index := 0; index+1 < len(values); index += 2 {
		x, errX := strconv.ParseFloat(values[index], 64)
		y, errY := strconv.ParseFloat(values[index+1], 64)
		if errX == nil && errY == nil {
			points = append(points, geometry.Point{X: x, Y: y})
		}
	}
	return points
}

func inheritedRole(role string) bool {
	switch role {
	case "background", "bridge-mask", "decoration", "label", "legend", "node", "reserved", "edge":
		return true
	default:
		return false
	}
}

func parseSVGNumber(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	n, err := strconv.ParseFloat(value, 64)
	return n, err == nil
}

func parseSVGPath(path string) ([]geometry.Point, error) {
	tokens := svgTokens(path)
	var points []geometry.Point
	var current geometry.Point
	var subpathStart geometry.Point
	var previousCubicControl geometry.Point
	hasPreviousCubic := false
	var previousQuadraticControl geometry.Point
	hasPreviousQuadratic := false
	command := byte(0)
	for len(tokens) > 0 {
		if len(tokens[0]) == 1 && strings.ContainsRune("MmLlHhVvCcSsQqTtAaZz", rune(tokens[0][0])) {
			command = tokens[0][0]
			tokens = tokens[1:]
		}
		if command == 0 {
			return nil, fmt.Errorf("unsupported SVG path")
		}
		next := func() (float64, error) {
			if len(tokens) == 0 {
				return 0, fmt.Errorf("missing coordinate")
			}
			n, err := strconv.ParseFloat(tokens[0], 64)
			tokens = tokens[1:]
			return n, err
		}
		switch command {
		case 'Z', 'z':
			if current != subpathStart {
				points = append(points, subpathStart)
				current = subpathStart
			}
			command = 0
			hasPreviousCubic = false
			hasPreviousQuadratic = false
		case 'M', 'm', 'L', 'l':
			x, e := next()
			if e != nil {
				return nil, e
			}
			y, e := next()
			if e != nil {
				return nil, e
			}
			if command == 'm' || command == 'l' {
				x += current.X
				y += current.Y
			}
			current = geometry.Point{X: x, Y: y}
			points = append(points, current)
			if command == 'M' || command == 'm' {
				subpathStart = current
			}
			if command == 'M' {
				command = 'L'
			}
			if command == 'm' {
				command = 'l'
			}
			hasPreviousCubic = false
			hasPreviousQuadratic = false
		case 'H', 'h':
			x, e := next()
			if e != nil {
				return nil, e
			}
			if command == 'h' {
				x += current.X
			}
			current.X = x
			points = append(points, current)
			hasPreviousCubic = false
			hasPreviousQuadratic = false
		case 'V', 'v':
			y, e := next()
			if e != nil {
				return nil, e
			}
			if command == 'v' {
				y += current.Y
			}
			current.Y = y
			points = append(points, current)
			hasPreviousCubic = false
			hasPreviousQuadratic = false
		case 'Q', 'q':
			controlX, e := next()
			if e != nil {
				return nil, e
			}
			controlY, e := next()
			if e != nil {
				return nil, e
			}
			endX, e := next()
			if e != nil {
				return nil, e
			}
			endY, e := next()
			if e != nil {
				return nil, e
			}
			if command == 'q' {
				controlX, controlY = controlX+current.X, controlY+current.Y
				endX, endY = endX+current.X, endY+current.Y
			}
			end := geometry.Point{X: endX, Y: endY}
			points = append(points, sampleQuadratic(current, geometry.Point{X: controlX, Y: controlY}, end, 12)...)
			current = end
			hasPreviousCubic = false
			previousQuadraticControl = geometry.Point{X: controlX, Y: controlY}
			hasPreviousQuadratic = true
		case 'C', 'c':
			firstX, e := next()
			if e != nil {
				return nil, e
			}
			firstY, e := next()
			if e != nil {
				return nil, e
			}
			secondX, e := next()
			if e != nil {
				return nil, e
			}
			secondY, e := next()
			if e != nil {
				return nil, e
			}
			endX, e := next()
			if e != nil {
				return nil, e
			}
			endY, e := next()
			if e != nil {
				return nil, e
			}
			if command == 'c' {
				firstX, firstY = firstX+current.X, firstY+current.Y
				secondX, secondY = secondX+current.X, secondY+current.Y
				endX, endY = endX+current.X, endY+current.Y
			}
			end := geometry.Point{X: endX, Y: endY}
			points = append(points, sampleCubic(current, geometry.Point{X: firstX, Y: firstY}, geometry.Point{X: secondX, Y: secondY}, end, 16)...)
			current = end
			previousCubicControl = geometry.Point{X: secondX, Y: secondY}
			hasPreviousCubic = true
			hasPreviousQuadratic = false
		case 'S', 's':
			secondX, e := next()
			if e != nil {
				return nil, e
			}
			secondY, e := next()
			if e != nil {
				return nil, e
			}
			endX, e := next()
			if e != nil {
				return nil, e
			}
			endY, e := next()
			if e != nil {
				return nil, e
			}
			if command == 's' {
				secondX, secondY = secondX+current.X, secondY+current.Y
				endX, endY = endX+current.X, endY+current.Y
			}
			first := current
			if hasPreviousCubic {
				first = geometry.Point{X: 2*current.X - previousCubicControl.X, Y: 2*current.Y - previousCubicControl.Y}
			}
			second := geometry.Point{X: secondX, Y: secondY}
			end := geometry.Point{X: endX, Y: endY}
			points = append(points, sampleCubic(current, first, second, end, 16)...)
			current = end
			previousCubicControl = second
			hasPreviousCubic = true
			hasPreviousQuadratic = false
		case 'T', 't':
			endX, e := next()
			if e != nil {
				return nil, e
			}
			endY, e := next()
			if e != nil {
				return nil, e
			}
			if command == 't' {
				endX, endY = endX+current.X, endY+current.Y
			}
			control := current
			if hasPreviousQuadratic {
				control = geometry.Point{X: 2*current.X - previousQuadraticControl.X, Y: 2*current.Y - previousQuadraticControl.Y}
			}
			end := geometry.Point{X: endX, Y: endY}
			points = append(points, sampleQuadratic(current, control, end, 12)...)
			current = end
			previousQuadraticControl = control
			hasPreviousQuadratic = true
			hasPreviousCubic = false
		case 'A', 'a':
			rx, e := next()
			if e != nil {
				return nil, e
			}
			ry, e := next()
			if e != nil {
				return nil, e
			}
			rotation, e := next()
			if e != nil {
				return nil, e
			}
			largeArc, e := next()
			if e != nil {
				return nil, e
			}
			sweep, e := next()
			if e != nil {
				return nil, e
			}
			endX, e := next()
			if e != nil {
				return nil, e
			}
			endY, e := next()
			if e != nil {
				return nil, e
			}
			if command == 'a' {
				endX, endY = endX+current.X, endY+current.Y
			}
			end := geometry.Point{X: endX, Y: endY}
			points = append(points, sampleArc(current, rx, ry, rotation, largeArc != 0, sweep != 0, end, 20)...)
			current = end
			hasPreviousCubic = false
			hasPreviousQuadratic = false
		default:
			return nil, fmt.Errorf("unsupported SVG path command %c", command)
		}
	}
	return points, nil
}

func svgTokens(value string) []string {
	var out []string
	for i := 0; i < len(value); {
		if unicode.IsSpace(rune(value[i])) || value[i] == ',' {
			i++
			continue
		}
		if strings.ContainsRune("MmLlHhVvCcSsQqTtAaZz", rune(value[i])) {
			out = append(out, value[i:i+1])
			i++
			continue
		}
		j := i + 1
		for j < len(value) && !unicode.IsSpace(rune(value[j])) && value[j] != ',' && !strings.ContainsRune("MmLlHhVvCcSsQqTtAaZz", rune(value[j])) {
			j++
		}
		out = append(out, value[i:j])
		i = j
	}
	return out
}

func sampleQuadratic(start, control, end geometry.Point, steps int) []geometry.Point {
	points := make([]geometry.Point, 0, steps)
	for index := 1; index <= steps; index++ {
		t := float64(index) / float64(steps)
		u := 1 - t
		points = append(points, geometry.Point{
			X: u*u*start.X + 2*u*t*control.X + t*t*end.X,
			Y: u*u*start.Y + 2*u*t*control.Y + t*t*end.Y,
		})
	}
	return points
}

func sampleCubic(start, first, second, end geometry.Point, steps int) []geometry.Point {
	points := make([]geometry.Point, 0, steps)
	for index := 1; index <= steps; index++ {
		t := float64(index) / float64(steps)
		u := 1 - t
		points = append(points, geometry.Point{
			X: u*u*u*start.X + 3*u*u*t*first.X + 3*u*t*t*second.X + t*t*t*end.X,
			Y: u*u*u*start.Y + 3*u*u*t*first.Y + 3*u*t*t*second.Y + t*t*t*end.Y,
		})
	}
	return points
}

func sampleArc(start geometry.Point, rx, ry, rotation float64, largeArc, sweep bool, end geometry.Point, steps int) []geometry.Point {
	rx, ry = math.Abs(rx), math.Abs(ry)
	if rx <= 1e-9 || ry <= 1e-9 || start == end {
		return []geometry.Point{end}
	}
	phi := math.Mod(rotation, 360) * math.Pi / 180
	cosPhi, sinPhi := math.Cos(phi), math.Sin(phi)
	dx, dy := (start.X-end.X)/2, (start.Y-end.Y)/2
	xPrime := cosPhi*dx + sinPhi*dy
	yPrime := -sinPhi*dx + cosPhi*dy
	scale := xPrime*xPrime/(rx*rx) + yPrime*yPrime/(ry*ry)
	if scale > 1 {
		factor := math.Sqrt(scale)
		rx, ry = rx*factor, ry*factor
	}
	numerator := math.Max(0, rx*rx*ry*ry-rx*rx*yPrime*yPrime-ry*ry*xPrime*xPrime)
	denominator := rx*rx*yPrime*yPrime + ry*ry*xPrime*xPrime
	coefficient := 0.0
	if denominator > 1e-12 {
		coefficient = math.Sqrt(numerator / denominator)
	}
	if largeArc == sweep {
		coefficient = -coefficient
	}
	centerXPrime := coefficient * (rx * yPrime / ry)
	centerYPrime := coefficient * (-ry * xPrime / rx)
	centerX := cosPhi*centerXPrime - sinPhi*centerYPrime + (start.X+end.X)/2
	centerY := sinPhi*centerXPrime + cosPhi*centerYPrime + (start.Y+end.Y)/2
	startAngle := math.Atan2((yPrime-centerYPrime)/ry, (xPrime-centerXPrime)/rx)
	endAngle := math.Atan2((-yPrime-centerYPrime)/ry, (-xPrime-centerXPrime)/rx)
	delta := endAngle - startAngle
	if sweep && delta < 0 {
		delta += 2 * math.Pi
	} else if !sweep && delta > 0 {
		delta -= 2 * math.Pi
	}
	points := make([]geometry.Point, 0, steps)
	for index := 1; index <= steps; index++ {
		angle := startAngle + delta*float64(index)/float64(steps)
		points = append(points, geometry.Point{
			X: centerX + rx*math.Cos(angle)*cosPhi - ry*math.Sin(angle)*sinPhi,
			Y: centerY + rx*math.Cos(angle)*sinPhi + ry*math.Sin(angle)*cosPhi,
		})
	}
	return points
}

func segmentHitsInterior(a, b geometry.Point, box geometry.Bounds) bool {
	dx, dy := b.X-a.X, b.Y-a.Y
	p := []float64{-dx, dx, -dy, dy}
	q := []float64{a.X - box.Left, box.Right - a.X, a.Y - box.Top, box.Bottom - a.Y}
	lo, hi := 0.0, 1.0
	for i := range p {
		if p[i] == 0 {
			if q[i] < 0 {
				return false
			}
			continue
		}
		r := q[i] / p[i]
		if p[i] < 0 {
			if r > hi {
				return false
			}
			if r > lo {
				lo = r
			}
		} else {
			if r < lo {
				return false
			}
			if r < hi {
				hi = r
			}
		}
	}
	if hi-lo <= geometry.EPSILON {
		return false
	}
	mid := geometry.Point{X: a.X + dx*(lo+hi)/2, Y: a.Y + dy*(lo+hi)/2}
	return geometry.PointInBounds(mid, box, 0, true)
}
