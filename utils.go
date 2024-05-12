package main

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

func abs[T signed](v T) T {
	if v < 0 {
		v = -v
	}

	if v < 0 {
		v -= 1 // minInt -> maxInt
	}

	return v
}
