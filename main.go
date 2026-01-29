// package main

// import (
// 	"InvisibleWaterMarkingSystem/Watermark"
// 	"fmt"
// 	"image"
// 	"image/jpeg"
// 	"os"
// )

// func main() {
// 	// ============ EMBEDDING ============
// 	fmt.Println("=== Watermark Embedding ===")

// 	// Load original image
// 	file, err := os.Open("Car.jpg")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	img, _, err := image.Decode(file)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Embed watermark
// 	message := "Hello World"
// 	fmt.Printf("Embedding message: \"%s\"\n", message)

// 	ycb := Watermark.Embed_Watermark(img, message)
// 	fmt.Println("✓ Watermark embedded successfully")

// 	// Save watermarked image
// 	outFile, err := os.Create("Watermarked_Image.jpg")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer outFile.Close()

// 	options := &jpeg.Options{
// 		Quality: 100, // 1–100
// 	}

// 	err = jpeg.Encode(outFile, ycb, options)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("✓ Watermarked image saved as 'Watermarked_Image.jpg'")

// 	// ============ EXTRACTION ============
// 	fmt.Println("=== Watermark Extraction ===")

// 	// Load watermarked image
// 	wmFile, err := os.Open("Watermarked_Image.jpg")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer wmFile.Close()

// 	wmImg, _, err := image.Decode(wmFile)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Extract watermark - simple mode
// 	fmt.Println("\n--- Simple Extraction ---")
// 	extractedMessage, err := Watermark.ExtractSingleMessage(wmImg)
// 	if err != nil {
// 		fmt.Printf("Extraction error: %v\n", err)
// 	} else {
// 		fmt.Printf("\nFinal extracted message: \"%s\"\n", extractedMessage)

// 		if extractedMessage == message {
// 			fmt.Println("✓ SUCCESS: Extracted message matches original!")
// 		} else {
// 			fmt.Printf("✗ MISMATCH: Original=\"%s\", Extracted=\"%s\"\n", message, extractedMessage)
// 		}
// 	}

// 	// Uncomment for detailed extraction information:
// 	// fmt.Println("\n--- Verbose Extraction ---")
// 	// Watermark.Extract_Watermark_Verbose(wmImg)
// }

package main

import (
	"InvisibleWaterMarkingSystem/Watermark"
	"fmt"
	"image"
	"image/jpeg"
	"os"
)

func main() {
	// Load original image
	file, err := os.Open("Car.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	message := "Hello World"

	// ============================================
	// TEST 1: Single Block Pipeline Test
	// ============================================
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  TEST 1: Single 8x8 Block Pipeline                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Create a simple test block
	testBlock := make([][]float64, 8)
	for i := range testBlock {
		testBlock[i] = make([]float64, 8)
		for j := range testBlock[i] {
			testBlock[i][j] = float64(i*8 + j) // Simple pattern
		}
	}

	// Test embedding bit pattern [1, 0]
	Watermark.TraceWatermarkPipeline(testBlock, 1, 0)

	// ============================================
	// TEST 2: Embed and Check in Y Matrix
	// ============================================
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  TEST 2: Full Embedding Process                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	fmt.Println("\n--- Embedding watermark ---")
	ycb := Watermark.Embed_Watermark(img, message)

	// Save watermarked image
	outFile, err := os.Create("Watermarked_Image.jpg")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	options := &jpeg.Options{Quality: 100}
	err = jpeg.Encode(outFile, ycb, options)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n✓ Watermarked image saved")

	// ============================================
	// TEST 3: Check Watermark in Y Matrix
	// ============================================
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  TEST 3: Check Y Matrix After Embedding                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Reconstruct Y matrix from watermarked image
	wmFile, err := os.Open("Watermarked_Image.jpg")
	if err != nil {
		panic(err)
	}
	defer wmFile.Close()

	wmImg, _, err := image.Decode(wmFile)
	if err != nil {
		panic(err)
	}

	_, wmYmatrix := Watermark.ConvertToYC(wmImg)

	// Check if watermark exists in Y matrix
	Watermark.CheckWatermarkInYMatrix(wmYmatrix, message)

	// ============================================
	// TEST 4: Check Single Tile
	// ============================================
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  TEST 4: Detailed Single Tile Analysis                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Get first tile from HL band
	img_DWT := Watermark.PerformCompleteDWT(wmYmatrix)

	if len(img_DWT.HL) >= 128 && len(img_DWT.HL[0]) >= 128 {
		// Extract first tile
		tile := make([][]float64, 128)
		for i := 0; i < 128; i++ {
			tile[i] = make([]float64, 128)
			copy(tile[i], img_DWT.HL[i][0:128])
		}

		Watermark.CheckWatermarkInTile(tile, message)
	} else {
		fmt.Println("⚠️  Image too small for tile analysis")
	}

	// ============================================
	// TEST 5: Check Single 8x8 Block
	// ============================================
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  TEST 5: Single 8x8 Block Analysis                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	if len(img_DWT.HL) >= 8 && len(img_DWT.HL[0]) >= 8 {
		// Extract first 8x8 block from first tile
		block := make([][]float64, 8)
		for i := 0; i < 8; i++ {
			block[i] = make([]float64, 8)
			copy(block[i], img_DWT.HL[i][0:8])
		}

		// First bits of the message
		stream := Watermark.BuildWatermarkBits(message)
		if len(stream) >= 2 {
			Watermark.CheckWatermarkInBlock(block, stream[0], stream[1])
		}
	}

	// ============================================
	// TEST 6: Normal Extraction
	// ============================================
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  TEST 6: Standard Extraction Process                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	extractedMessage, err := Watermark.ExtractSingleMessage(wmImg)
	if err != nil {
		fmt.Printf("\n✗ Extraction failed: %v\n", err)
	} else {
		fmt.Printf("\n✓ Extracted message: \"%s\"\n", extractedMessage)
		if extractedMessage == message {
			fmt.Println("✓ SUCCESS: Message matches original!")
		} else {
			fmt.Printf("✗ MISMATCH: Expected \"%s\", got \"%s\"\n", message, extractedMessage)
		}
	}

	// Summary
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  DIAGNOSTIC SUMMARY                                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("\nReview the tests above to identify where the watermark is lost:")
	fmt.Println("  - Test 1: Verifies DCT→Embed→IDCT→Extract pipeline")
	fmt.Println("  - Test 2: Shows embedding process details")
	fmt.Println("  - Test 3: Checks if watermark survives IDWT")
	fmt.Println("  - Test 4: Analyzes single tile bit-by-bit")
	fmt.Println("  - Test 5: Checks single 8x8 block coefficients")
	fmt.Println("  - Test 6: Standard extraction process")
}
