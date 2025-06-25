package util

func Chunks[T any](slice []T, chunkSize int) [][]T {
	if len(slice) == 0 {
		return nil
	}

	divided := make([][]T, (len(slice)+chunkSize-1)/chunkSize)
	prev := 0
	index := 0
	till := len(slice) - chunkSize

	for prev < till {
		next := prev + chunkSize
		divided[index] = slice[prev:next]
		prev = next
		index++
	}

	divided[index] = slice[prev:]

	return divided
}
