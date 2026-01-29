package Watermark

func bytesToBits(data []byte) []int {
	bits := make([]int, 0, len(data)*8)

	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1
			bits = append(bits, int(bit))
		}
	}
	return bits
}

func bitsToBytes(bits []int) []byte {
	var result []byte
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			b = (b << 1) | byte(bits[i+j])
		}
		result = append(result, b)
	}
	return result
}

// BuildWatermarkBits converts a message string to a watermark bit stream with flags
func BuildWatermarkBits(message string) []int {
	startFlag := []int{
		1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 0, 0, 0, 0,
	}
	endFlag := []int{
		0, 0, 0, 0, 1, 1, 1, 1,
		0, 0, 0, 0, 1, 1, 1, 1,
	}

	dataBits := bytesToBits([]byte(message))

	stream := make([]int, 0)
	stream = append(stream, startFlag...)
	stream = append(stream, dataBits...)
	stream = append(stream, endFlag...)

	return stream
}
