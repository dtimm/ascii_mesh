package main

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

const tol = 1e-9

func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

func TestBuildEdges_SingleTet(t *testing.T) {
	got := BuildEdges([]Tet{{1, 2, 3, 4}})
	want := []Edge{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuildEdges_TwoTetsDeduplication(t *testing.T) {
	got := BuildEdges([]Tet{{1, 2, 3, 4}, {2, 3, 4, 5}})
	// 6 + 6 = 12, minus 3 shared edges (1-2, 1-3, 2-3 in 0-based) = 9
	if len(got) != 9 {
		t.Errorf("got %d edges, want 9", len(got))
	}
}

func TestBuildEdges_Empty(t *testing.T) {
	got := BuildEdges([]Tet{})
	if len(got) != 0 {
		t.Errorf("got %d edges, want 0", len(got))
	}
}

func TestBuildEdges_Sorted(t *testing.T) {
	got := BuildEdges([]Tet{{1, 2, 3, 4}, {2, 3, 4, 5}})
	for i := 1; i < len(got); i++ {
		if got[i][0] < got[i-1][0] || (got[i][0] == got[i-1][0] && got[i][1] <= got[i-1][1]) {
			t.Errorf("edges not sorted at index %d: %v before %v", i, got[i-1], got[i])
		}
	}
}

func TestProjectPoints_Origin(t *testing.T) {
	got := ProjectPoints([]Point3D{{0, 0, 0}})
	if !approxEqual(got[0].U, 0, tol) || !approxEqual(got[0].V, 0, tol) || !approxEqual(got[0].Depth, 0, tol) {
		t.Errorf("got %+v, want {0, 0, 0}", got[0])
	}
}

func TestProjectPoints_UnitX(t *testing.T) {
	got := ProjectPoints([]Point3D{{1, 0, 0}})
	// e1 = (-1, 0, 0), so u = dot((1,0,0), (-1,0,0)) = -1
	if !approxEqual(got[0].U, -1, tol) {
		t.Errorf("U = %f, want -1", got[0].U)
	}
}

func TestProjectPoints_UnitY(t *testing.T) {
	got := ProjectPoints([]Point3D{{0, 1, 0}})
	s := 1.0 / math.Sqrt(2)
	if !approxEqual(got[0].V, s, tol) {
		t.Errorf("V = %f, want %f", got[0].V, s)
	}
}

func TestProjectPoints_UnitZ(t *testing.T) {
	got := ProjectPoints([]Point3D{{0, 0, 1}})
	s := 1.0 / math.Sqrt(2)
	if !approxEqual(got[0].Depth, -s, tol) {
		t.Errorf("Depth = %f, want %f", got[0].Depth, -s)
	}
}

func TestNormalizeToCanvas_TwoPoints(t *testing.T) {
	us := []float64{0, 1}
	vs := []float64{0, 1}
	xs, ys := NormalizeToCanvas(us, vs, 11, 11, 0)
	// (0,0) -> screen (0, 10) because Y flips; (1,1) -> screen (10, 0)
	if xs[0] != 0 || ys[0] != 10 {
		t.Errorf("point 0: got (%d,%d), want (0,10)", xs[0], ys[0])
	}
	if xs[1] != 10 || ys[1] != 0 {
		t.Errorf("point 1: got (%d,%d), want (10,0)", xs[1], ys[1])
	}
}

func TestNormalizeToCanvas_WithMargin(t *testing.T) {
	us := []float64{0, 1}
	vs := []float64{0, 1}
	xs, ys := NormalizeToCanvas(us, vs, 11, 11, 1)
	// margin=1: range maps to [1, 9]
	if xs[0] != 1 || ys[0] != 9 {
		t.Errorf("point 0: got (%d,%d), want (1,9)", xs[0], ys[0])
	}
	if xs[1] != 9 || ys[1] != 1 {
		t.Errorf("point 1: got (%d,%d), want (9,1)", xs[1], ys[1])
	}
}

func TestNormalizeToCanvas_DegenerateURange(t *testing.T) {
	us := []float64{5, 5}
	vs := []float64{0, 1}
	xs, _ := NormalizeToCanvas(us, vs, 11, 11, 0)
	// All same u, should not panic. Both x values should be equal.
	if xs[0] != xs[1] {
		t.Errorf("expected same x for degenerate u range, got %d and %d", xs[0], xs[1])
	}
}

func TestBuildDepthBuffer_EmptyTets(t *testing.T) {
	buf := BuildDepthBuffer([]int{}, []int{}, []float64{}, []Tet{}, 5, 5)
	if !math.IsInf(buf[2][2], 1) {
		t.Errorf("expected +Inf for empty tets, got %f", buf[2][2])
	}
}

func TestBuildDepthBuffer_DegenerateTriangle(t *testing.T) {
	// All 4 nodes are collinear in screen space (same x, different y)
	xs := []int{2, 2, 2, 2}
	ys := []int{0, 1, 2, 3}
	depths := []float64{0, 0, 0, 0}
	tets := []Tet{{1, 2, 3, 4}}
	buf := BuildDepthBuffer(xs, ys, depths, tets, 5, 5)
	// Degenerate triangles should be skipped
	if !math.IsInf(buf[0][0], 1) {
		t.Errorf("expected +Inf for degenerate, got %f", buf[0][0])
	}
}

func TestBuildDepthBuffer_SingleTetInsidePixel(t *testing.T) {
	// A tet that forms a visible triangle on screen
	xs := []int{0, 4, 2, 2}
	ys := []int{0, 0, 4, 2}
	depths := []float64{1.0, 1.0, 1.0, 1.0}
	tets := []Tet{{1, 2, 3, 4}}
	buf := BuildDepthBuffer(xs, ys, depths, tets, 5, 5)
	// Pixel (2, 1) should be inside the face (0,0)-(4,0)-(2,2)
	if math.IsInf(buf[1][2], 1) {
		t.Error("expected finite depth inside triangle, got +Inf")
	}
}

func TestBuildDepthBuffer_NearerFaceWins(t *testing.T) {
	// Two overlapping triangles at different depths
	xs := []int{0, 4, 2, 0, 4, 2}
	ys := []int{0, 0, 4, 0, 0, 4}
	depths := []float64{5.0, 5.0, 5.0, 1.0, 1.0, 1.0}
	// Two tets that cover similar area. Second is nearer.
	tets := []Tet{{1, 2, 3, 3}, {4, 5, 6, 6}}
	buf := BuildDepthBuffer(xs, ys, depths, tets, 5, 5)
	// Check a pixel inside both triangles
	if buf[1][2] > 2.0 {
		t.Errorf("expected nearer depth ~1.0, got %f", buf[1][2])
	}
}

func makeCanvas(w, h int) [][]byte {
	c := make([][]byte, h)
	for i := range c {
		c[i] = make([]byte, w)
		for j := range c[i] {
			c[i][j] = ' '
		}
	}
	return c
}

func makeInfBuf(w, h int) [][]float64 {
	buf := make([][]float64, h)
	for i := range buf {
		buf[i] = make([]float64, w)
		for j := range buf[i] {
			buf[i][j] = math.Inf(1)
		}
	}
	return buf
}

func TestDrawEdge_SinglePointVisible(t *testing.T) {
	canvas := makeCanvas(5, 5)
	buf := makeInfBuf(5, 5)
	DrawEdgeHiddenLine(canvas, 2, 2, 0.0, 2, 2, 0.0, buf, 1e-6)
	if canvas[2][2] != '*' {
		t.Errorf("got %c, want *", canvas[2][2])
	}
}

func TestDrawEdge_HorizontalLineVisible(t *testing.T) {
	canvas := makeCanvas(5, 1)
	buf := makeInfBuf(5, 1)
	DrawEdgeHiddenLine(canvas, 0, 0, 0.0, 4, 0, 0.0, buf, 1e-6)
	for x := 0; x < 5; x++ {
		if canvas[0][x] != '*' {
			t.Errorf("canvas[0][%d] = %c, want *", x, canvas[0][x])
		}
	}
}

func TestDrawEdge_HiddenPoint(t *testing.T) {
	canvas := makeCanvas(5, 5)
	buf := makeInfBuf(5, 5)
	buf[2][2] = 0.0 // face at depth 0
	// Edge point at depth 5 (behind the face)
	DrawEdgeHiddenLine(canvas, 2, 2, 5.0, 2, 2, 5.0, buf, 1e-6)
	if canvas[2][2] != '.' {
		t.Errorf("got %c, want .", canvas[2][2])
	}
}

func TestDrawEdge_VisibleNotOverwrittenByHidden(t *testing.T) {
	canvas := makeCanvas(5, 5)
	buf := makeInfBuf(5, 5)
	// First draw visible
	DrawEdgeHiddenLine(canvas, 2, 2, 0.0, 2, 2, 0.0, buf, 1e-6)
	// Now set depth buffer to make next draw hidden
	buf[2][2] = 0.0
	DrawEdgeHiddenLine(canvas, 2, 2, 5.0, 2, 2, 5.0, buf, 1e-6)
	if canvas[2][2] != '*' {
		t.Errorf("got %c, want * (should not be overwritten)", canvas[2][2])
	}
}

func TestDrawEdge_OutOfBounds(t *testing.T) {
	canvas := makeCanvas(3, 3)
	buf := makeInfBuf(3, 3)
	// Should not panic even though line extends outside
	DrawEdgeHiddenLine(canvas, -1, -1, 0.0, 5, 5, 0.0, buf, 1e-6)
}

func TestParseMesh_Valid(t *testing.T) {
	input := `# sample mesh
nodes
0.0  0.0  0.0
1.0  0.0  2.0
2.0  1.0  0.0

tets
1 2 3 1
`
	elems, err := ParseMesh(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems.NodeCoordinates) != 3 {
		t.Errorf("got %d nodes, want 3", len(elems.NodeCoordinates))
	}
}

func TestParseMesh_ValidNodeValues(t *testing.T) {
	input := "nodes\n1.5 -2.0 3.0\n\ntets\n1 1 1 1\n"
	elems, err := ParseMesh(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Point3D{1.5, -2.0, 3.0}
	if elems.NodeCoordinates[0] != want {
		t.Errorf("got %v, want %v", elems.NodeCoordinates[0], want)
	}
}

func TestParseMesh_ValidTetValues(t *testing.T) {
	input := "nodes\n0 0 0\n\ntets\n1 2 3 4\n"
	elems, err := ParseMesh(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Tet{1, 2, 3, 4}
	if elems.Tets[0] != want {
		t.Errorf("got %v, want %v", elems.Tets[0], want)
	}
}

func TestParseMesh_CommentsAndBlanks(t *testing.T) {
	input := `
# this is a comment

nodes
# another comment
0 0 0
1 1 1

# comment between sections

tets

# comment in tets
1 2 1 2
`
	elems, err := ParseMesh(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems.NodeCoordinates) != 2 {
		t.Errorf("got %d nodes, want 2", len(elems.NodeCoordinates))
	}
}

func TestParseMesh_MissingNodes(t *testing.T) {
	input := "tets\n1 2 3 4\n"
	_, err := ParseMesh(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for missing nodes section")
	}
}

func TestParseMesh_MissingTets(t *testing.T) {
	input := "nodes\n0 0 0\n"
	_, err := ParseMesh(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for missing tets section")
	}
}

func TestParseMesh_BadNodeLine(t *testing.T) {
	input := "nodes\n1.0 2.0\n\ntets\n1 2 3 4\n"
	_, err := ParseMesh(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for node line with 2 fields")
	}
}

func TestParseMesh_BadTetLine(t *testing.T) {
	input := "nodes\n0 0 0\n\ntets\n1 2 3\n"
	_, err := ParseMesh(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for tet line with 3 fields")
	}
}

func TestParseMesh_InvalidFloat(t *testing.T) {
	input := "nodes\nabc 0 0\n\ntets\n1 1 1 1\n"
	_, err := ParseMesh(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for non-numeric node coordinate")
	}
}

func TestParseMesh_InvalidInt(t *testing.T) {
	input := "nodes\n0 0 0\n\ntets\n1 abc 3 4\n"
	_, err := ParseMesh(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for non-integer tet index")
	}
}

var sampleElements = Elements{
	NodeCoordinates: []Point3D{
		{0.0, 0.0, 0.0},
		{1.0, 0.0, 2.0},
		{2.0, 1.0, 0.0},
		{2.0, -1.0, 0.0},
		{3.0, 0.0, 2.0},
		{4.0, 0.0, 0.0},
	},
	Tets: []Tet{
		{1, 2, 3, 4},
		{2, 3, 4, 5},
		{3, 4, 5, 6},
	},
}

func TestASCIIArt_NodeLabelsPresent(t *testing.T) {
	art := ASCIIArtFromElements(sampleElements, 50, 15)
	for _, label := range "ABCDEF" {
		if !strings.Contains(art, string(label)) {
			t.Errorf("missing node label %c", label)
		}
	}
}

func TestASCIIArt_Dimensions(t *testing.T) {
	art := ASCIIArtFromElements(sampleElements, 50, 15)
	lines := strings.Split(art, "\n")
	if len(lines) != 15 {
		t.Errorf("got %d lines, want 15", len(lines))
	}
	for i, line := range lines {
		if len(line) != 50 {
			t.Errorf("line %d has %d chars, want 50", i, len(line))
		}
	}
}

func TestASCIIArt_ContainsVisibleEdges(t *testing.T) {
	art := ASCIIArtFromElements(sampleElements, 50, 15)
	if !strings.Contains(art, "*") {
		t.Error("output should contain visible edge chars '*'")
	}
}

func TestASCIIArt_ContainsHiddenEdges(t *testing.T) {
	art := ASCIIArtFromElements(sampleElements, 50, 15)
	if !strings.Contains(art, ".") {
		t.Error("output should contain hidden edge chars '.'")
	}
}

func TestASCIIArt_GoldenOutput(t *testing.T) {
	art := ASCIIArtFromElements(sampleElements, 50, 15)
	want := "" +
		"                                                  \n" +
		"            E***********************B             \n" +
		"          ** *...               ...* **           \n" +
		"         *    *  ...         ...  *    *          \n" +
		"        *      *    ...   ...    *      **        \n" +
		"      **        *     ..C..     *         *       \n" +
		"     *          .*....  .  ....*..         **     \n" +
		"    *     ......  *     .     *   .....      *    \n" +
		"  **......         *    .    *         .......**  \n" +
		" F..                *   .   *                 ..A \n" +
		"    ***...           *  .  *           ...****    \n" +
		"          ***...      * . *       ..***           \n" +
		"                ***... *.* ...****                \n" +
		"                      **D**                       \n" +
		"                                                  "
	if art != want {
		t.Errorf("output mismatch.\ngot:\n%s\n\nwant:\n%s", art, want)
	}
}
