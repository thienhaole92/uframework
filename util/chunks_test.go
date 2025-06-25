package util_test

import (
	"reflect"
	"testing"

	"github.com/thienhaole92/uframework/util"
)

func TestChunks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     []int
		chunkSize int
		expected  [][]int
	}{
		{
			name:      "Empty slice",
			input:     []int{},
			chunkSize: 2,
			expected:  nil,
		},
		{
			name:      "Single chunk",
			input:     []int{1, 2, 3},
			chunkSize: 3,
			expected:  [][]int{{1, 2, 3}},
		},
		{
			name:      "Multiple chunks of equal size",
			input:     []int{1, 2, 3, 4, 5, 6},
			chunkSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}, {5, 6}},
		},
		{
			name:      "Multiple chunks with last chunk smaller",
			input:     []int{1, 2, 3, 4, 5},
			chunkSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name:      "Chunk size larger than slice length",
			input:     []int{1, 2, 3},
			chunkSize: 5,
			expected:  [][]int{{1, 2, 3}},
		},
		{
			name:      "Nil slice",
			input:     nil,
			chunkSize: 2,
			expected:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := util.Chunks(test.input, test.chunkSize)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Chunks() = %v, want %v", result, test.expected)
			}
		})
	}
}
