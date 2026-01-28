package Watermark

import (
	"fmt"
	"image"
	"math"
	"sync"
	"time"
)

func ConvertToYC(img image.Image) (*image.YCbCr, [][]float64) {
	bounds := img.Bounds()

	// Create a new YCbCr image with the same size
	ycb := image.NewYCbCr(bounds, image.YCbCrSubsampleRatio444)

	Ymatrix := make([][]float64, bounds.Dy())
	for i := range Ymatrix {
		Ymatrix[i] = make([]float64, bounds.Dx())
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert using the standard Go formula
			rr := float64(r >> 8)
			gg := float64(g >> 8)
			bb := float64(b >> 8)

			Y := 0.299*rr + 0.587*gg + 0.114*bb
			Cb := -0.1687*rr - 0.3313*gg + 0.5*bb + 128
			Cr := 0.5*rr - 0.4187*gg - 0.0813*bb + 128

			yi := y - bounds.Min.Y
			xi := x - bounds.Min.X

			Ymatrix[yi][xi] = float64(Y) / 255.0
			ycb.Y[ycb.YOffset(x, y)] = uint8(Y)
			ycb.Cb[ycb.COffset(x, y)] = uint8(Cb)
			ycb.Cr[ycb.COffset(x, y)] = uint8(Cr)
		}
	}
	return ycb, Ymatrix
}

func haarL(input []float64, min float64, max float64) ([]float64, float64, float64) {
	n := len(input)
	output := make([]float64, n/2)

	for i := 0; i < n/2; i++ {
		a := (input[2*i] + input[2*i+1]) / math.Sqrt2
		if a < min {
			min = a
		}
		if a > max {
			max = a
		}
		output[i] = a
	}
	return output, min, max
}

func haarH(input []float64, min float64, max float64) ([]float64, float64, float64) {
	n := len(input)
	output := make([]float64, n/2)

	for i := 0; i < n/2; i++ {
		d := (input[2*i] - input[2*i+1]) / math.Sqrt2
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
		output[i] = d
	}
	return output, min, max
}

// Parallel version of CreateLLLH using goroutines
func CreateLLLH(Ymatrix [][]float64) ([][]float64, [][]float64, []float64) {
	diff := make([]float64, 2)
	h := len(Ymatrix)
	w := len(Ymatrix[0])

	// Row-wise DWT (L) - Parallel processing
	temp := make([][]float64, h)
	var wg sync.WaitGroup

	for i := 0; i < h; i++ {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			temp[row], _, _ = haarL(Ymatrix[row], 0, 0)
		}(i)
	}
	wg.Wait()

	// Find min/max for L transform
	Lmin := temp[0][0]
	Lmax := temp[0][0]
	var minMaxMutex sync.Mutex

	// Column-wise DWT (L) - Parallel processing
	result_1 := make([][]float64, h/2)
	for i := range result_1 {
		result_1[i] = make([]float64, w/2)
	}

	for col := 0; col < w/2; col++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			column := make([]float64, h)
			for row := 0; row < h; row++ {
				column[row] = temp[row][c]
			}

			transformed, min, max := haarL(column, Lmin, Lmax)

			// Update global min/max with mutex
			minMaxMutex.Lock()
			if min < Lmin {
				Lmin = min
			}
			if max > Lmax {
				Lmax = max
			}
			minMaxMutex.Unlock()

			for row := 0; row < h/2; row++ {
				result_1[row][c] = transformed[row]
			}
		}(col)
	}
	wg.Wait()

	fmt.Println("Minimum and Maximum of L : ", Lmin, Lmax)
	diff[0] = Lmax - Lmin

	// Column-wise DWT (H) - Parallel processing
	Hmin := temp[0][0]
	Hmax := temp[0][0]

	result_2 := make([][]float64, h/2)
	for i := range result_2 {
		result_2[i] = make([]float64, w/2)
	}

	for col := 0; col < w/2; col++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			column := make([]float64, h)
			for row := 0; row < h; row++ {
				column[row] = temp[row][c]
			}

			transformed, min, max := haarH(column, Hmin, Hmax)

			// Update global min/max with mutex
			minMaxMutex.Lock()
			if min < Hmin {
				Hmin = min
			}
			if max > Hmax {
				Hmax = max
			}
			minMaxMutex.Unlock()

			for row := 0; row < h/2; row++ {
				result_2[row][c] = transformed[row]
			}
		}(col)
	}
	wg.Wait()

	fmt.Println("Minimum and Maximum of H : ", Hmin, Hmax)
	diff[1] = Hmax - Hmin

	return result_1, result_2, diff
}

// Parallel version of CreateHLHH using goroutines
// func CreateHLHH(Ymatrix [][]float64) ([][]float64, [][]float64, []float64) {
// 	diff := make([]float64, 2)
// 	h := len(Ymatrix)
// 	w := len(Ymatrix[0])

// 	// Row-wise DWT (H) - Parallel processing
// 	temp := make([][]float64, h)
// 	var wg sync.WaitGroup

// 	for i := 0; i < h; i++ {
// 		wg.Add(1)
// 		go func(row int) {
// 			defer wg.Done()
// 			temp[row], _, _ = haarH(Ymatrix[row], 0, 0)
// 		}(i)
// 	}
// 	wg.Wait()

// 	// Find min/max for L transform
// 	Lmin := temp[0][0]
// 	Lmax := temp[0][0]
// 	var minMaxMutex sync.Mutex

// 	// Column-wise DWT (L) - Parallel processing
// 	result_1 := make([][]float64, h/2)
// 	for i := range result_1 {
// 		result_1[i] = make([]float64, w/2)
// 	}

// 	for col := 0; col < w/2; col++ {
// 		wg.Add(1)
// 		go func(c int) {
// 			defer wg.Done()
// 			column := make([]float64, h)
// 			for row := 0; row < h; row++ {
// 				column[row] = temp[row][c]
// 			}

// 			transformed, min, max := haarL(column, Lmin, Lmax)

// 			// Update global min/max with mutex
// 			minMaxMutex.Lock()
// 			if min < Lmin {
// 				Lmin = min
// 			}
// 			if max > Lmax {
// 				Lmax = max
// 			}
// 			minMaxMutex.Unlock()

// 			for row := 0; row < h/2; row++ {
// 				result_1[row][c] = transformed[row]
// 			}
// 		}(col)
// 	}
// 	wg.Wait()

// 	fmt.Println("Minimum and Maximum of L : ", Lmin, Lmax)
// 	diff[0] = Lmax - Lmin

// 	// Column-wise DWT (H) - Parallel processing
// 	Hmin := temp[0][0]
// 	Hmax := temp[0][0]

// 	result_2 := make([][]float64, h/2)
// 	for i := range result_2 {
// 		result_2[i] = make([]float64, w/2)
// 	}

// 	for col := 0; col < w/2; col++ {
// 		wg.Add(1)
// 		go func(c int) {
// 			defer wg.Done()
// 			column := make([]float64, h)
// 			for row := 0; row < h; row++ {
// 				column[row] = temp[row][c]
// 			}

// 			transformed, min, max := haarH(column, Hmin, Hmax)

// 			// Update global min/max with mutex
// 			minMaxMutex.Lock()
// 			if min < Hmin {
// 				Hmin = min
// 			}
// 			if max > Hmax {
// 				Hmax = max
// 			}
// 			minMaxMutex.Unlock()

// 			for row := 0; row < h/2; row++ {
// 				result_2[row][c] = transformed[row]
// 			}
// 		}(col)
// 	}
// 	wg.Wait()

// 	fmt.Println("Minimum and Maximum of H : ", Hmin, Hmax)
// 	diff[1] = Hmax - Hmin

// 	return result_1, result_2, diff
// }

// Convert matrix back to image
// func YToImage(matrix [][]float64, diff float64) *image.Gray {
// 	h := len(matrix)
// 	w := len(matrix[0])
// 	img := image.NewGray(image.Rect(0, 0, w, h))

// 	for y := 0; y < h; y++ {
// 		for x := 0; x < w; x++ {
// 			val := uint8(math.Min(math.Max((matrix[y][x]*255), 0), 255))
// 			img.SetGray(x, y, color.Gray{Y: val})
// 		}
// 	}
// 	return img
// }

// func SaveImage(filename string, img image.Image) {
// 	outFile, err := os.Create(filename)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer outFile.Close()

// 	err = jpeg.Encode(outFile, img, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func PerformWaveletTransform(img image.Image) [][]float64 {
	

	// Convert to YCbCr and get Y matrix
	_, Ymatrix := ConvertToYC(img)

	t1 := time.Now()

	// Use goroutines to process LL/LH and HL/HH in parallel
	// var wg sync.WaitGroup
	// var LL, LH [][]float64
	// var diffL []float64

	// wg.Add(2)

	// Process LL and LH in parallel
	// go func() {
	// 	defer wg.Done()
	// 	LL, LH, diffL = CreateLLLH(Ymatrix)
	// }()
	_, LH, _ := CreateLLLH(Ymatrix)
	// Process HL and HH in parallel
	// go func() {
	// 	defer wg.Done()
	// 	HL, HH, diffH = CreateHLHH(Ymatrix)
	// }()

	//wg.Wait()

	t2 := time.Now()
	fmt.Println("Time taken = ", t2.Sub(t1))

	// Convert matrices back to images
	// LLimg := YToImage(LL, diffL[0])
	// LHimg := YToImage(LH, diffL[1])
	// HLimg := YToImage(HL, diffH[0])
	// HHimg := YToImage(HH, diffH[1])

	// Save images
	// SaveImage("LL_DWT2.jpg", LLimg)
	// SaveImage("LH_DWT2.jpg", LHimg)
	// SaveImage("HL_DWT2.jpg", HLimg)
	// SaveImage("HH_DWT2.jpg", HHimg)

	return  LH
}
