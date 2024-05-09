package main

import "math"

func sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func cos32(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

func sin32(x float32) float32 {
	return float32(math.Sin(float64(x)))
}
