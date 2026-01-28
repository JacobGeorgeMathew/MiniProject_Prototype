package main

import (
	"InvisibleWaterMarkingSystem/Watermark"
	"fmt"
	"image"
	"math"
	"os"
)

func main() {
	// Load image
	file, err := os.Open("Car.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	img_DWT := Watermark.PerformWaveletTransform(img)
	fmt.Println("Converted to DWT : ")
	temp := make([][]float64, 8)
	for i := 0; i < 8; i++ {
		temp[i] = make([]float64, 8)
	}

	h := len(img_DWT)
	w := len(img_DWT[0])
	for i := 0; i < int(math.Floor(float64(h)/128)); i += 128 {
		for j := 0; j < int(math.Floor(float64(w)/128)); j += 128 {

		}
	}
	//Watermark.PerformEmbedd(img)
}
