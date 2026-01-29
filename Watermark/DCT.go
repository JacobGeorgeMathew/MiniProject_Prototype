package Watermark

import (
	"math"
)

func qimEmbed(c float64, bit int, delta float64) float64 {
	base := math.Floor(c/delta) * delta

	if bit == 0 {
		return base + delta/4
	}
	return base + 3*delta/4
}

// FIXED qimExtract - handles negative coefficients correctly
func qimExtract(c float64, delta float64) int {
	// Fix for negative coefficients
	// Instead of math.Mod, compute remainder properly
	base := math.Floor(c/delta) * delta
	remainder := c - base

	// Now remainder is always in [0, delta)
	if remainder < delta/2 {
		return 0
	}
	return 1
}

func dct1D(input []float64) []float64 {
	N := len(input)
	output := make([]float64, N)

	for k := 0; k < N; k++ {
		sum := 0.0
		for n := 0; n < N; n++ {
			sum += input[n] * math.Cos(
				math.Pi*float64(2*n+1)*float64(k)/(2*float64(N)),
			)
		}

		alpha := math.Sqrt(2.0 / float64(N))
		if k == 0 {
			alpha = math.Sqrt(1.0 / float64(N))
		}
		output[k] = alpha * sum
	}
	return output
}

func dct2D(block [][]float64) [][]float64 {
	N := len(block)

	// Row-wise DCT
	temp := make([][]float64, N)
	for i := 0; i < N; i++ {
		temp[i] = dct1D(block[i])
	}

	// Column-wise DCT
	result := make([][]float64, N)
	for i := range result {
		result[i] = make([]float64, N)
	}

	for j := 0; j < N; j++ {
		col := make([]float64, N)
		for i := 0; i < N; i++ {
			col[i] = temp[i][j]
		}
		colDCT := dct1D(col)
		for i := 0; i < N; i++ {
			result[i][j] = colDCT[i]
		}
	}

	return result
}

// PerformEmbedd modifies the block in-place by embedding watermark bits
func PerformEmbedd(block [][]float64, bits []int) {
	// Use alpha = 10.0 for stronger watermark
	alpha := 10.0

	// Perform DCT
	dctBlock := dct2D(block)

	// Embed watermark in mid-frequency coefficients
	dctBlock[1][3] = qimEmbed(dctBlock[1][3], bits[0], alpha)
	dctBlock[3][1] = qimEmbed(dctBlock[3][1], bits[1], alpha)

	// Perform IDCT and copy back to original block
	idctBlock := idct2D(dctBlock)
	for i := 0; i < len(block); i++ {
		copy(block[i], idctBlock[i])
	}
}

func PerformExtract(block [][]float64) []int {
	// Must match the alpha used in PerformEmbedd
	alpha := 10.0

	dctBlock := dct2D(block)

	bits := make([]int, 2)

	bits[0] = qimExtract(dctBlock[1][3], alpha)
	bits[1] = qimExtract(dctBlock[3][1], alpha)

	return bits
}
