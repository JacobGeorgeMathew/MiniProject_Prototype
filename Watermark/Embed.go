package Watermark

import (
	"fmt"
	"image"
	"math"
)

func getBlock(matrix [][]float64, x, y, B int) [][]float64 {
	block := make([][]float64, B)
	for i := 0; i < B; i++ {
		block[i] = make([]float64, B)
		copy(block[i], matrix[y+i][x:x+B])
	}
	return block
}

func putBlock(matrix [][]float64, block [][]float64, x, y int) {
	for i := 0; i < len(block); i++ {
		copy(matrix[y+i][x:x+len(block)], block[i])
	}
}

func embed_in_a_tile(tile [][]float64, stream []int) [][]float64 {
	bitIndex := 0
	for by := 0; by < 128 && bitIndex < len(stream)-1; by += 8 {
		for bx := 0; bx < 128 && bitIndex < len(stream)-1; bx += 8 {

			block := getBlock(tile, bx, by, 8)
			bits := make([]int, 2)
			bits[0] = stream[bitIndex]
			bits[1] = stream[bitIndex+1]

			// PerformEmbedd now handles DCT and IDCT internally
			PerformEmbedd(block, bits)

			// Block is already in spatial domain, just put it back
			putBlock(tile, block, bx, by)

			bitIndex += 2
		}
	}
	return tile
}

func Embed_Watermark(img image.Image, message string) *image.YCbCr {

	stream := buildWatermarkBits(message)

	ycb, Ymatrix := ConvertToYC(img)

	img_DWT := PerformCompleteDWT(Ymatrix)

	fmt.Println("Converted to DWT")

	h := len(img_DWT.HL)
	w := len(img_DWT.HL[0])

	// Process tiles
	for i := 0; i < int(math.Floor(float64(h)/128)); i++ {
		for j := 0; j < int(math.Floor(float64(w)/128)); j++ {
			block := getBlock(img_DWT.HL, j*128, i*128, 128)

			tile := embed_in_a_tile(block, stream)

			putBlock(img_DWT.HL, tile, j*128, i*128)
		}
	}

	Ymatrix = PerformCompleteIDWTFromResult(img_DWT)
	Modify_YComponent(ycb, Ymatrix)
	return ycb
}
