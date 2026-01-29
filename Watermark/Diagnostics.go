package Watermark

import (
	"fmt"
	"math"
)

// =====================================================
// DIAGNOSTIC TOOL 1: Check Watermark in Y Matrix
// =====================================================

// CheckWatermarkInYMatrix checks if watermark exists in Y matrix after IDWT
func CheckWatermarkInYMatrix(Ymatrix [][]float64, message string) {
	fmt.Println("\n=== Checking Watermark in Y Matrix (After IDWT) ===")

	stream := BuildWatermarkBits(message)

	// Perform DWT to get to frequency domain
	img_DWT := PerformCompleteDWT(Ymatrix)

	h := len(img_DWT.HL)
	w := len(img_DWT.HL[0])

	numTilesY := int(math.Floor(float64(h) / 128))
	numTilesX := int(math.Floor(float64(w) / 128))

	fmt.Printf("HL band size: %dx%d\n", w, h)
	fmt.Printf("Number of tiles: %dx%d = %d\n", numTilesX, numTilesY, numTilesX*numTilesY)
	fmt.Printf("Expected watermark bits: %d\n\n", len(stream))

	tilesWithWatermark := 0

	for i := 0; i < numTilesY; i++ {
		for j := 0; j < numTilesX; j++ {
			tile := getBlock(img_DWT.HL, j*128, i*128, 128)

			// Extract bits from this tile
			extractedBits := extractBitsFromTile(tile)

			// Check if message exists
			extractedMsg, found := findMessage(extractedBits)

			if found {
				tilesWithWatermark++
				fmt.Printf("  Tile [%d,%d]: ✓ Message found: \"%s\"\n", i, j, extractedMsg)
			} else {
				fmt.Printf("  Tile [%d,%d]: ✗ No valid message\n", i, j)
				// Show first 64 bits for debugging
				fmt.Print("    First 64 bits: ")
				for k := 0; k < 64 && k < len(extractedBits); k++ {
					fmt.Printf("%d", extractedBits[k])
					if (k+1)%8 == 0 {
						fmt.Print(" ")
					}
				}
				fmt.Println()
			}
		}
	}

	fmt.Printf("\nSummary: %d/%d tiles contain valid watermark\n", tilesWithWatermark, numTilesX*numTilesY)

	if tilesWithWatermark == 0 {
		fmt.Println("⚠️  WARNING: No watermark found in Y matrix after IDWT!")
		fmt.Println("   Problem is likely in: DWT → Embed → IDWT pipeline")
	} else if tilesWithWatermark < numTilesX*numTilesY {
		fmt.Println("⚠️  WARNING: Watermark found in some but not all tiles")
		fmt.Println("   Problem might be in: Tile iteration during embedding")
	} else {
		fmt.Println("✓ SUCCESS: Watermark found in all tiles!")
	}
}

// extractBitsFromTile extracts watermark bits from a single tile
func extractBitsFromTile(tile [][]float64) []int {
	var extractedBits []int

	for by := 0; by < 128; by += 8 {
		for bx := 0; bx < 128; bx += 8 {
			block := getBlock(tile, bx, by, 8)

			// Extract 2 bits from this block
			bits := PerformExtract(block)
			extractedBits = append(extractedBits, bits...)
		}
	}

	return extractedBits
}

// =====================================================
// DIAGNOSTIC TOOL 2: Check Watermark in Single Tile
// =====================================================

// CheckWatermarkInTile checks if watermark exists in a tile (HL band)
func CheckWatermarkInTile(tile [][]float64, expectedMessage string) {
	fmt.Println("\n=== Checking Watermark in Single Tile ===")

	stream := BuildWatermarkBits(expectedMessage)
	fmt.Printf("Expected message: \"%s\"\n", expectedMessage)
	fmt.Printf("Expected bit stream length: %d bits\n", len(stream))

	// Extract bits from tile
	extractedBits := extractBitsFromTile(tile)
	fmt.Printf("Extracted bits from tile: %d bits\n", len(extractedBits))

	// Show first 128 bits
	fmt.Println("\nFirst 128 extracted bits:")
	for i := 0; i < 128 && i < len(extractedBits); i++ {
		fmt.Printf("%d", extractedBits[i])
		if (i+1)%8 == 0 {
			fmt.Print(" ")
		}
		if (i+1)%32 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()

	// Show expected bits
	fmt.Println("\nExpected bit stream (first 128 bits):")
	for i := 0; i < 128 && i < len(stream); i++ {
		fmt.Printf("%d", stream[i])
		if (i+1)%8 == 0 {
			fmt.Print(" ")
		}
		if (i+1)%32 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()

	// Compare bit by bit
	matchCount := 0
	mismatchCount := 0
	compareLength := len(stream)
	if len(extractedBits) < compareLength {
		compareLength = len(extractedBits)
	}

	for i := 0; i < compareLength; i++ {
		if extractedBits[i] == stream[i] {
			matchCount++
		} else {
			mismatchCount++
		}
	}

	accuracy := float64(matchCount) / float64(compareLength) * 100.0
	fmt.Printf("\nBit accuracy: %d/%d (%.2f%%)\n", matchCount, compareLength, accuracy)

	// Try to extract message
	extractedMsg, found := findMessage(extractedBits)

	if found {
		fmt.Printf("\n✓ Message extracted: \"%s\"\n", extractedMsg)
		if extractedMsg == expectedMessage {
			fmt.Println("✓ SUCCESS: Extracted message matches expected!")
		} else {
			fmt.Printf("✗ MISMATCH: Expected \"%s\", got \"%s\"\n", expectedMessage, extractedMsg)
		}
	} else {
		fmt.Println("\n✗ FAILED: Could not extract valid message (flags not found)")

		// Detailed flag analysis
		analyzeFlags(extractedBits, stream)
	}
}

// analyzeFlags checks if start/end flags are present
func analyzeFlags(extractedBits []int, expectedBits []int) {
	startFlag := []int{
		1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 0, 0, 0, 0,
	}
	endFlag := []int{
		0, 0, 0, 0, 1, 1, 1, 1,
		0, 0, 0, 0, 1, 1, 1, 1,
	}

	fmt.Println("\n--- Flag Analysis ---")

	// Check start flag in extracted bits
	fmt.Print("Start flag in extracted: ")
	startFound := false
	for i := 0; i <= len(extractedBits)-len(startFlag); i++ {
		match := true
		for j := 0; j < len(startFlag); j++ {
			if extractedBits[i+j] != startFlag[j] {
				match = false
				break
			}
		}
		if match {
			fmt.Printf("Found at position %d ✓\n", i)
			startFound = true
			break
		}
	}
	if !startFound {
		fmt.Println("NOT FOUND ✗")
		// Show where it should be
		fmt.Print("  Expected at position 0: ")
		for i := 0; i < len(startFlag) && i < len(extractedBits); i++ {
			if extractedBits[i] == startFlag[i] {
				fmt.Printf("\033[32m%d\033[0m", extractedBits[i])
			} else {
				fmt.Printf("\033[31m%d\033[0m", extractedBits[i])
			}
		}
		fmt.Println()
	}

	// Check end flag
	fmt.Print("End flag in extracted: ")
	endFound := false
	expectedEndPos := len(expectedBits) - len(endFlag)
	for i := 0; i <= len(extractedBits)-len(endFlag); i++ {
		match := true
		for j := 0; j < len(endFlag); j++ {
			if extractedBits[i+j] != endFlag[j] {
				match = false
				break
			}
		}
		if match {
			fmt.Printf("Found at position %d", i)
			if i == expectedEndPos {
				fmt.Println(" (correct position) ✓")
			} else {
				fmt.Printf(" (expected at %d) ⚠️\n", expectedEndPos)
			}
			endFound = true
			break
		}
	}
	if !endFound {
		fmt.Println("NOT FOUND ✗")
	}
}

// =====================================================
// DIAGNOSTIC TOOL 3: Check Single 8x8 Block After IDCT
// =====================================================

// CheckWatermarkInBlock checks if watermark bits are preserved in a single 8x8 block
func CheckWatermarkInBlock(block [][]float64, bit0 int, bit1 int) {
	fmt.Println("\n=== Checking Watermark in 8x8 Block (After IDCT) ===")

	fmt.Printf("Expected bits to embed: [%d, %d]\n", bit0, bit1)

	// First, show the block values
	fmt.Println("\nBlock values (spatial domain):")
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Printf("%7.2f ", block[i][j])
		}
		fmt.Println()
	}

	// Perform DCT
	dctBlock := dct2D(block)

	fmt.Println("\nDCT coefficients:")
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			fmt.Printf("%7.2f ", dctBlock[i][j])
		}
		fmt.Println()
	}

	// Extract bits
	alpha := 10.0 // Should match your embedding alpha
	extractedBit0 := qimExtract(dctBlock[1][3], alpha)
	extractedBit1 := qimExtract(dctBlock[3][1], alpha)

	fmt.Printf("\nWatermark coefficients:\n")
	fmt.Printf("  Position [1][3]: %.4f\n", dctBlock[1][3])
	fmt.Printf("  Position [3][1]: %.4f\n", dctBlock[3][1])

	fmt.Printf("\nExtracted bits: [%d, %d]\n", extractedBit0, extractedBit1)

	if extractedBit0 == bit0 && extractedBit1 == bit1 {
		fmt.Println("✓ SUCCESS: Watermark bits correctly preserved!")
	} else {
		fmt.Println("✗ FAILED: Watermark bits lost or corrupted!")
		fmt.Printf("  Expected: [%d, %d]\n", bit0, bit1)
		fmt.Printf("  Got:      [%d, %d]\n", extractedBit0, extractedBit1)

		// Detailed QIM analysis
		fmt.Println("\nQIM Analysis:")
		analyzeQIM(dctBlock[1][3], bit0, alpha, "[1][3]")
		analyzeQIM(dctBlock[3][1], bit1, alpha, "[3][1]")
	}
}

// analyzeQIM analyzes QIM quantization for a coefficient
func analyzeQIM(coefficient float64, expectedBit int, delta float64, position string) {
	fmt.Printf("\n  Position %s:\n", position)
	fmt.Printf("    Coefficient value: %.4f\n", coefficient)
	fmt.Printf("    Delta (alpha): %.4f\n", delta)

	base := math.Floor(coefficient/delta) * delta
	fmt.Printf("    Quantization base: %.4f\n", base)

	remainder := math.Mod(coefficient, delta)
	fmt.Printf("    Remainder: %.4f\n", remainder)

	extractedBit := 0
	if remainder >= delta/2 {
		extractedBit = 1
	}

	fmt.Printf("    Expected bit: %d\n", expectedBit)
	fmt.Printf("    Extracted bit: %d\n", extractedBit)

	if expectedBit == 0 {
		expectedRange := fmt.Sprintf("[%.2f, %.2f)", base, base+delta/2)
		fmt.Printf("    Expected range for bit 0: %s\n", expectedRange)
	} else {
		expectedRange := fmt.Sprintf("[%.2f, %.2f)", base+delta/2, base+delta)
		fmt.Printf("    Expected range for bit 1: %s\n", expectedRange)
	}

	if extractedBit == expectedBit {
		fmt.Println("    ✓ Bit correctly preserved")
	} else {
		fmt.Println("    ✗ Bit corrupted")
	}
}

// =====================================================
// DIAGNOSTIC TOOL 4: Trace Full Pipeline
// =====================================================

// TraceWatermarkPipeline traces watermark through entire pipeline
func TraceWatermarkPipeline(originalBlock [][]float64, bit0 int, bit1 int) {
	fmt.Println("\n=== Tracing Watermark Through Pipeline ===")

	alpha := 10.0

	// Step 1: Original block
	fmt.Println("\n[Step 1] Original 8x8 block (spatial domain)")
	fmt.Printf("First row: ")
	for j := 0; j < 8; j++ {
		fmt.Printf("%.2f ", originalBlock[0][j])
	}
	fmt.Println()

	// Step 2: DCT
	dctBlock := dct2D(originalBlock)
	fmt.Println("\n[Step 2] After DCT (frequency domain)")
	fmt.Printf("Coefficient [1][3] = %.4f\n", dctBlock[1][3])
	fmt.Printf("Coefficient [3][1] = %.4f\n", dctBlock[3][1])

	// Step 3: Embed watermark
	embeddedBlock := make([][]float64, 8)
	for i := range embeddedBlock {
		embeddedBlock[i] = make([]float64, 8)
		copy(embeddedBlock[i], dctBlock[i])
	}

	embeddedBlock[1][3] = qimEmbed(dctBlock[1][3], bit0, alpha)
	embeddedBlock[3][1] = qimEmbed(dctBlock[3][1], bit1, alpha)

	fmt.Println("\n[Step 3] After QIM embedding")
	fmt.Printf("Coefficient [1][3] = %.4f (was %.4f, embedded bit %d)\n",
		embeddedBlock[1][3], dctBlock[1][3], bit0)
	fmt.Printf("Coefficient [3][1] = %.4f (was %.4f, embedded bit %d)\n",
		embeddedBlock[3][1], dctBlock[3][1], bit1)

	// Step 4: IDCT
	spatialBlock := idct2D(embeddedBlock)
	fmt.Println("\n[Step 4] After IDCT (back to spatial domain)")
	fmt.Printf("First row: ")
	for j := 0; j < 8; j++ {
		fmt.Printf("%.2f ", spatialBlock[0][j])
	}
	fmt.Println()

	// Step 5: Extract
	extractDCT := dct2D(spatialBlock)
	extractedBit0 := qimExtract(extractDCT[1][3], alpha)
	extractedBit1 := qimExtract(extractDCT[3][1], alpha)

	fmt.Println("\n[Step 5] Extraction")
	fmt.Printf("Re-DCT coefficient [1][3] = %.4f\n", extractDCT[1][3])
	fmt.Printf("Re-DCT coefficient [3][1] = %.4f\n", extractDCT[3][1])
	fmt.Printf("Extracted bits: [%d, %d]\n", extractedBit0, extractedBit1)

	// Verification
	fmt.Println("\n[Verification]")
	if extractedBit0 == bit0 && extractedBit1 == bit1 {
		fmt.Println("✓ SUCCESS: DCT→Embed→IDCT→DCT→Extract pipeline works!")
	} else {
		fmt.Println("✗ FAILED: Pipeline corrupted the watermark")
		fmt.Printf("  Input bits:     [%d, %d]\n", bit0, bit1)
		fmt.Printf("  Extracted bits: [%d, %d]\n", extractedBit0, extractedBit1)
	}
}
