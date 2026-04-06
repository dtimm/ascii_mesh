# ascii_mesh

A CLI tool that renders tetrahedral meshes as ASCII art with hidden-line removal.

```
            E***********************B
          ** *...               ...* **
         *    *  ...         ...  *    *
        *      *    ...   ...    *      **
      **        *     ..C..     *         *
     *          .*....  .  ....*..         **
    *     ......  *     .     *   .....      *
  **......         *    .    *         .......**
 F..                *   .   *                 ..A
    ***...           *  .  *           ...****
          ***...      * . *       ..***
                ***... *.* ...****
                      **D**
```

## Usage

```
ascii_mesh <mesh-file>
```

```sh
go run . example.mesh
```

## Mesh file format

Mesh files are plain text with two sections: `nodes` and `tets`. Lines starting with `#` are comments. Blank lines are ignored.

```
# Example tetrahedral mesh
# 6 nodes, 3 tetrahedra

nodes
0.0   0.0   0.0
1.0   0.0   2.0
2.0   1.0   0.0
2.0  -1.0   0.0
3.0   0.0   2.0
4.0   0.0   0.0

tets
1 2 3 4
2 3 4 5
3 4 5 6
```

- **nodes** -- one line per node, three whitespace-separated coordinates (x y z)
- **tets** -- one line per tetrahedron, four whitespace-separated 1-based node indices
- Sections can appear in any order

## Rendering

The tool uses an orthographic isometric projection (view direction along +Y, -Z) and a per-pixel depth buffer built from the triangular faces of each tetrahedron.

- `*` -- visible edge
- `.` -- hidden (occluded) edge
- `A`-`Z` -- node labels

## Building and testing

```sh
go build -o ascii_mesh .
go test -v -cover ./...
```
