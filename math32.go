package main

import "math"

func sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x))) // translates to SQRTSS on x86
}
