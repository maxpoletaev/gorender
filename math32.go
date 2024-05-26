package main

import (
	"math"

	"github.com/orsinium-labs/tinymath"
)

const (
	pi32 = float32(math.Pi)
)

var (
	// Faster but less accurate trig
	sin32 = tinymath.Sin
	cos32 = tinymath.Cos
	tan32 = tinymath.Tan
)

func sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x))) // translates to SQRTSS on x86
}
