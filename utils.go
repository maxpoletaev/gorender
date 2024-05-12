package main

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func abs[T integer](v T) T {
	if v < 0 {
		return -v
	}
	return v
}
