package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
)

const NodeLabels = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Point3D represents a 3D coordinate.
type Point3D [3]float64

// Edge is a pair of 0-based node indices, sorted (Lo < Hi).
type Edge [2]int

// Tet holds four 1-based node indices for a tetrahedron.
type Tet [4]int

// Elements holds the mesh data.
type Elements struct {
	NodeCoordinates []Point3D
	Tets            []Tet
}

// ProjectedPoint holds screen coordinates and depth.
type ProjectedPoint struct {
	U, V, Depth float64
}

// ParseMesh reads a mesh file from r and returns the parsed Elements.
// The format has "nodes" and "tets" section headers.
// Lines starting with # are comments; blank lines are ignored.
func ParseMesh(r io.Reader) (Elements, error) {
	var elems Elements
	var section string
	hasNodes := false
	hasTets := false

	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line == "nodes" {
			section = "nodes"
			hasNodes = true
			continue
		}
		if line == "tets" {
			section = "tets"
			hasTets = true
			continue
		}

		fields := strings.Fields(line)

		switch section {
		case "nodes":
			if len(fields) != 3 {
				return elems, fmt.Errorf("line %d: expected 3 coordinates, got %d", lineNum, len(fields))
			}
			var p Point3D
			for i, f := range fields {
				v, err := strconv.ParseFloat(f, 64)
				if err != nil {
					return elems, fmt.Errorf("line %d: invalid coordinate %q: %w", lineNum, f, err)
				}
				p[i] = v
			}
			elems.NodeCoordinates = append(elems.NodeCoordinates, p)

		case "tets":
			if len(fields) != 4 {
				return elems, fmt.Errorf("line %d: expected 4 node indices, got %d", lineNum, len(fields))
			}
			var tet Tet
			for i, f := range fields {
				v, err := strconv.Atoi(f)
				if err != nil {
					return elems, fmt.Errorf("line %d: invalid index %q: %w", lineNum, f, err)
				}
				tet[i] = v
			}
			elems.Tets = append(elems.Tets, tet)

		default:
			return elems, fmt.Errorf("line %d: data before any section header", lineNum)
		}
	}

	if err := scanner.Err(); err != nil {
		return elems, err
	}
	if !hasNodes {
		return elems, fmt.Errorf("missing 'nodes' section")
	}
	if !hasTets {
		return elems, fmt.Errorf("missing 'tets' section")
	}

	return elems, nil
}

// BuildEdges extracts unique edges from tetrahedra.
// Input tets use 1-based node indices; returned edges are 0-based and sorted.
func BuildEdges(tets []Tet) []Edge {
	seen := make(map[Edge]bool)
	for _, tet := range tets {
		idxs := [4]int{tet[0] - 1, tet[1] - 1, tet[2] - 1, tet[3] - 1}
		for i := 0; i < 4; i++ {
			for j := i + 1; j < 4; j++ {
				a, b := idxs[i], idxs[j]
				if a > b {
					a, b = b, a
				}
				seen[Edge{a, b}] = true
			}
		}
	}

	edges := make([]Edge, 0, len(seen))
	for e := range seen {
		edges = append(edges, e)
	}
	slices.SortFunc(edges, func(a, b Edge) int {
		if a[0] != b[0] {
			return a[0] - b[0]
		}
		return a[1] - b[1]
	})
	return edges
}

func dot3(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func cross3(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func normalize3(v [3]float64) [3]float64 {
	n := math.Sqrt(dot3(v, v))
	return [3]float64{v[0] / n, v[1] / n, v[2] / n}
}

// ProjectPoints performs an orthographic isometric projection.
// View direction is along normalized(0, +1, -1), world up is (0, 0, 1).
// Returns U (screen x), V (screen y), and Depth for each point.
func ProjectPoints(points []Point3D) []ProjectedPoint {
	d := normalize3([3]float64{0, 1, -1})
	worldUp := [3]float64{0, 0, 1}

	e1 := normalize3(cross3(worldUp, d))
	e2 := normalize3(cross3(d, e1))

	result := make([]ProjectedPoint, len(points))
	for i, p := range points {
		result[i] = ProjectedPoint{
			U:     dot3([3]float64(p), e1),
			V:     dot3([3]float64(p), e2),
			Depth: dot3([3]float64(p), d),
		}
	}
	return result
}

// NormalizeToCanvas maps 2D coordinates to integer canvas coordinates.
// Uses independent scaling (fills width and height). Flips Y for screen coords.
func NormalizeToCanvas(us, vs []float64, width, height, margin int) (xs, ys []int) {
	umin, umax := us[0], us[0]
	vmin, vmax := vs[0], vs[0]
	for _, u := range us {
		umin = min(umin, u)
		umax = max(umax, u)
	}
	for _, v := range vs {
		vmin = min(vmin, v)
		vmax = max(vmax, v)
	}

	urange := umax - umin
	vrange := vmax - vmin
	if urange == 0 {
		urange = 1.0
	}
	if vrange == 0 {
		vrange = 1.0
	}

	scaleX := float64(width-1-2*margin) / urange
	scaleY := float64(height-1-2*margin) / vrange

	xs = make([]int, len(us))
	ys = make([]int, len(vs))
	for i := range us {
		xs[i] = int((us[i] - umin) * scaleX + float64(margin))
		ys[i] = (height - 1) - int((vs[i]-vmin)*scaleY+float64(margin))
	}
	return xs, ys
}

// BuildDepthBuffer rasterizes triangular faces from tetrahedra into a depth buffer
// using barycentric interpolation. Each tet produces 4 triangular faces.
func BuildDepthBuffer(xs, ys []int, depths []float64, tets []Tet, width, height int) [][]float64 {
	buf := make([][]float64, height)
	for i := range buf {
		buf[i] = make([]float64, width)
		for j := range buf[i] {
			buf[i][j] = math.Inf(1)
		}
	}

	type face struct{ a, b, c int }
	var faces []face
	for _, tet := range tets {
		a, b, c, d := tet[0]-1, tet[1]-1, tet[2]-1, tet[3]-1
		faces = append(faces, face{a, b, c}, face{a, b, d}, face{a, c, d}, face{b, c, d})
	}

	for _, f := range faces {
		x0, y0, z0 := xs[f.a], ys[f.a], depths[f.a]
		x1, y1, z1 := xs[f.b], ys[f.b], depths[f.b]
		x2, y2, z2 := xs[f.c], ys[f.c], depths[f.c]

		minX := max(min(x0, x1, x2), 0)
		maxX := min(max(x0, x1, x2), width-1)
		minY := max(min(y0, y1, y2), 0)
		maxY := min(max(y0, y1, y2), height-1)

		den := float64((y1-y2)*(x0-x2) + (x2-x1)*(y0-y2))
		if den == 0 {
			continue
		}

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				a := float64((y1-y2)*(x-x2)+(x2-x1)*(y-y2)) / den
				b := float64((y2-y0)*(x-x2)+(x0-x2)*(y-y2)) / den
				c := 1.0 - a - b

				if a >= 0 && b >= 0 && c >= 0 {
					depth := a*z0 + b*z1 + c*z2
					if depth < buf[y][x] {
						buf[y][x] = depth
					}
				}
			}
		}
	}

	return buf
}

// DrawEdgeHiddenLine draws an edge on the canvas with hidden-line style.
// '*' for visible points, '.' for hidden. Visible never overwritten by hidden.
func DrawEdgeHiddenLine(canvas [][]byte, x0, y0 int, z0 float64, x1, y1 int, z1 float64, depthBuf [][]float64, eps float64) {
	height := len(canvas)
	width := len(canvas[0])

	dx := x1 - x0
	dy := y1 - y0

	adx := dx
	if adx < 0 {
		adx = -adx
	}
	ady := dy
	if ady < 0 {
		ady = -ady
	}
	steps := max(adx, ady)

	if steps == 0 {
		if x0 >= 0 && x0 < width && y0 >= 0 && y0 < height {
			b := depthBuf[y0][x0]
			if math.IsInf(b, 1) || z0 <= b+eps {
				canvas[y0][x0] = '*'
			} else if canvas[y0][x0] == ' ' {
				canvas[y0][x0] = '.'
			}
		}
		return
	}

	for k := 0; k <= steps; k++ {
		t := float64(k) / float64(steps)
		x := int(math.RoundToEven(float64(x0) + float64(dx)*t))
		y := int(math.RoundToEven(float64(y0) + float64(dy)*t))

		if x >= 0 && x < width && y >= 0 && y < height {
			depth := z0 + t*(z1-z0)
			b := depthBuf[y][x]

			if math.IsInf(b, 1) || depth <= b+eps {
				canvas[y][x] = '*'
			} else if canvas[y][x] == ' ' {
				canvas[y][x] = '.'
			}
		}
	}
}

// ASCIIArtFromElements builds an ASCII art representation of the tetrahedral mesh
// with hidden-line rendering. Nodes are labeled A, B, C, ...
func ASCIIArtFromElements(elements Elements, width, height int) string {
	pts3d := elements.NodeCoordinates
	tets := elements.Tets

	edges := BuildEdges(tets)
	projected := ProjectPoints(pts3d)

	us := make([]float64, len(projected))
	vs := make([]float64, len(projected))
	depths := make([]float64, len(projected))
	for i, p := range projected {
		us[i] = p.U
		vs[i] = p.V
		depths[i] = p.Depth
	}

	xs, ys := NormalizeToCanvas(us, vs, width, height, 1)
	depthBuf := BuildDepthBuffer(xs, ys, depths, tets, width, height)

	canvas := make([][]byte, height)
	for i := range canvas {
		canvas[i] = make([]byte, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	for _, e := range edges {
		DrawEdgeHiddenLine(canvas,
			xs[e[0]], ys[e[0]], depths[e[0]],
			xs[e[1]], ys[e[1]], depths[e[1]],
			depthBuf, 1e-6)
	}

	for idx := range xs {
		x, y := xs[idx], ys[idx]
		if x >= 0 && x < width && y >= 0 && y < height {
			canvas[y][x] = NodeLabels[idx]
		}
	}

	lines := make([]string, height)
	for i, row := range canvas {
		lines[i] = string(row)
	}
	return strings.Join(lines, "\n")
}
