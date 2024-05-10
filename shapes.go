package main

func NewCube() *Mesh {
	vertices := []Vec3{
		{1, 1, -1},
		{1, -1, -1},
		{1, 1, 1},
		{1, -1, 1},
		{-1, 1, -1},
		{-1, -1, -1},
		{-1, 1, 1},
		{-1, -1, 1},
	}

	faces := []Face{
		{4, 2, 0, UV{}, UV{}, UV{}},
		{2, 7, 3, UV{}, UV{}, UV{}},
		{6, 5, 7, UV{}, UV{}, UV{}},
		{1, 7, 5, UV{}, UV{}, UV{}},
		{0, 3, 1, UV{}, UV{}, UV{}},
		{4, 1, 5, UV{}, UV{}, UV{}},
		{4, 6, 2, UV{}, UV{}, UV{}},
		{2, 6, 7, UV{}, UV{}, UV{}},
		{6, 4, 5, UV{}, UV{}, UV{}},
		{1, 3, 7, UV{}, UV{}, UV{}},
		{0, 2, 3, UV{}, UV{}, UV{}},
		{4, 0, 1, UV{}, UV{}, UV{}},
	}

	return NewMesh(vertices, faces)
}
