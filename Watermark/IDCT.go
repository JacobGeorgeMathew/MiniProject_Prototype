package Watermark

import "math"

func idct1D(input []float64) []float64 {
	N := len(input)
	output := make([]float64, N)

	for n := 0; n < N; n++ {
		sum := 0.0
		for k := 0; k < N; k++ {
			alpha := math.Sqrt(2.0 / float64(N))
			if k == 0 {
				alpha = math.Sqrt(1.0 / float64(N))
			}

			sum += alpha * input[k] *
				math.Cos(
					math.Pi*float64(2*n+1)*float64(k)/
						(2*float64(N)),
				)
		}
		output[n] = sum
	}
	return output
}

func idct2D(block [][]float64) [][]float64 {
	N := len(block)

	// Column-wise IDCT
	temp := make([][]float64, N)
	for i := range temp {
		temp[i] = make([]float64, N)
	}

	for j := 0; j < N; j++ {
		col := make([]float64, N)
		for i := 0; i < N; i++ {
			col[i] = block[i][j]
		}
		colIDCT := idct1D(col)
		for i := 0; i < N; i++ {
			temp[i][j] = colIDCT[i]
		}
	}

	// Row-wise IDCT
	result := make([][]float64, N)
	for i := 0; i < N; i++ {
		result[i] = idct1D(temp[i])
	}

	return result
}

func PerformIDCT(block [][]float64) {

}
