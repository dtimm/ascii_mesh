package main

import "fmt"

func main() {
	elements := Elements{
		NodeCoordinates: []Point3D{
			{0.0, 0.0, 0.0},  // A
			{1.0, 0.0, 2.0},  // B
			{2.0, 1.0, 0.0},  // C
			{2.0, -1.0, 0.0}, // D
			{3.0, 0.0, 2.0},  // E
			{4.0, 0.0, 0.0},  // F
		},
		Tets: []Tet{
			{1, 2, 3, 4},
			{2, 3, 4, 5},
			{3, 4, 5, 6},
		},
	}

	fmt.Println(ASCIIArtFromElements(elements, 50, 15))
}
