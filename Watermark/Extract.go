package Watermark

import (
	"fmt"
	"image"
)

// extractFromTile extracts watermark bits from a 128x128 tile
func extractFromTile(tile [][]float64) []int {
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

// findMessage locates the message between start and end flags
func findMessage(bits []int) (string, bool) {
	startFlag := []int{
		1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 0, 0, 0, 0,
	}
	endFlag := []int{
		0, 0, 0, 0, 1, 1, 1, 1,
		0, 0, 0, 0, 1, 1, 1, 1,
	}

	// Find start flag
	startIndex := -1
	for i := 0; i <= len(bits)-len(startFlag); i++ {
		match := true
		for j := 0; j < len(startFlag); j++ {
			if bits[i+j] != startFlag[j] {
				match = false
				break
			}
		}
		if match {
			startIndex = i + len(startFlag)
			break
		}
	}

	if startIndex == -1 {
		return "", false
	}

	// Find end flag after start
	endIndex := -1
	for i := startIndex; i <= len(bits)-len(endFlag); i++ {
		match := true
		for j := 0; j < len(endFlag); j++ {
			if bits[i+j] != endFlag[j] {
				match = false
				break
			}
		}
		if match {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return "", false
	}

	// Extract message bits between flags
	messageBits := bits[startIndex:endIndex]

	// Convert bits to bytes
	if len(messageBits)%8 != 0 {
		// Pad with zeros if needed
		padding := 8 - (len(messageBits) % 8)
		messageBits = append(messageBits, make([]int, padding)...)
	}

	messageBytes := bitsToBytes(messageBits)
	return string(messageBytes), true
}

// Extract_Watermark extracts the watermark message from a watermarked image
func Extract_Watermark(img image.Image) []string {
	// Convert image to YCbCr and get Y matrix
	_, Ymatrix := ConvertToYC(img)

	// Perform DWT
	img_DWT := PerformCompleteDWT(Ymatrix)

	fmt.Println("DWT completed for extraction")

	h := len(img_DWT.HL)
	w := len(img_DWT.HL[0])

	var messages []string
	tileCount := 0

	// Process each 128x128 tile
	numTilesY := h / 128
	numTilesX := w / 128

	fmt.Printf("Processing %d x %d = %d tiles\n", numTilesY, numTilesX, numTilesY*numTilesX)

	for i := 0; i < numTilesY; i++ {
		for j := 0; j < numTilesX; j++ {
			tileCount++

			// Get the tile
			tile := getBlock(img_DWT.HL, j*128, i*128, 128)

			// Extract bits from this tile
			extractedBits := extractFromTile(tile)

			// Try to find the message
			message, found := findMessage(extractedBits)

			if found {
				fmt.Printf("Tile [%d,%d] (tile #%d): Message found: \"%s\"\n", i, j, tileCount, message)
				messages = append(messages, message)
			} else {
				fmt.Printf("Tile [%d,%d] (tile #%d): No valid message found\n", i, j, tileCount)
			}
		}
	}

	return messages
}

// Extract_Watermark_Verbose provides detailed extraction information
func Extract_Watermark_Verbose(img image.Image) {
	_, Ymatrix := ConvertToYC(img)
	img_DWT := PerformCompleteDWT(Ymatrix)

	h := len(img_DWT.HL)
	w := len(img_DWT.HL[0])

	numTilesY := h / 128
	numTilesX := w / 128

	fmt.Println("\n=== Watermark Extraction (Verbose Mode) ===")
	fmt.Printf("Image size: %dx%d\n", w*2, h*2)
	fmt.Printf("HL band size: %dx%d\n", w, h)
	fmt.Printf("Number of tiles: %d x %d = %d\n\n", numTilesY, numTilesX, numTilesY*numTilesX)

	for i := 0; i < numTilesY; i++ {
		for j := 0; j < numTilesX; j++ {
			fmt.Printf("--- Tile [%d,%d] ---\n", i, j)

			tile := getBlock(img_DWT.HL, j*128, i*128, 128)
			extractedBits := extractFromTile(tile)

			fmt.Printf("Extracted %d bits from tile\n", len(extractedBits))

			// Show first 32 bits
			fmt.Print("First 32 bits: ")
			for k := 0; k < 32 && k < len(extractedBits); k++ {
				fmt.Printf("%d", extractedBits[k])
			}
			fmt.Println()

			message, found := findMessage(extractedBits)

			if found {
				fmt.Printf("✓ Message found: \"%s\"\n", message)
			} else {
				fmt.Println("✗ No valid message found (flags not detected)")
			}
			fmt.Println()
		}
	}
}

// ExtractSingleMessage attempts to extract one consistent message across all tiles
func ExtractSingleMessage(img image.Image) (string, error) {
	messages := Extract_Watermark(img)

	if len(messages) == 0 {
		return "", fmt.Errorf("no watermark found in any tile")
	}

	// Check if all messages are the same
	firstMessage := messages[0]
	allSame := true

	for _, msg := range messages {
		if msg != firstMessage {
			allSame = false
			break
		}
	}

	if allSame {
		fmt.Printf("\n✓ Consistent message found in %d/%d tiles: \"%s\"\n",
			len(messages), len(messages), firstMessage)
		return firstMessage, nil
	} else {
		fmt.Printf("\n⚠ Warning: Found different messages in tiles\n")
		fmt.Println("Messages found:")
		messageCount := make(map[string]int)
		for _, msg := range messages {
			messageCount[msg]++
		}
		for msg, count := range messageCount {
			fmt.Printf("  \"%s\": %d tiles\n", msg, count)
		}

		// Return the most common message
		maxCount := 0
		mostCommon := ""
		for msg, count := range messageCount {
			if count > maxCount {
				maxCount = count
				mostCommon = msg
			}
		}

		return mostCommon, nil
	}
}
