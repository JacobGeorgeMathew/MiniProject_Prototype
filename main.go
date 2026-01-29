package main

import (
	"InvisibleWaterMarkingSystem/Watermark"
	"fmt"
	"image"
	"image/jpeg"
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

	ycb := Watermark.Embed_Watermark(img, "Hello World")
	fmt.Println("Watermark embedded successfully")

	outFile, err := os.Create("Watermarked_Image.jpg")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	options := &jpeg.Options{
		Quality: 100, // 1–100
	}

	err = jpeg.Encode(outFile, ycb, options)

	if err != nil {
		panic(err)
	}
	//Watermark.PerformEmbedd(img)
}
