package parser

import (
	"fmt"
	"testing"
)

func TestParseChain(t *testing.T) {
	examples := []string{
		"chain a * 3 b",
		"chain mrchou * 2 b * 3",
		"chain a a",
		"chain [a b ] * 4 b",
		"chain a*2 b",
		"chain [a * 2 b] * 4",
		"chain one * 2 two",
	}
	expected := [][]string{
		[]string{"a", "a", "a", "b"},
		[]string{"mrchou", "mrchou", "b", "b", "b"},
		[]string{"a", "a"},
		[]string{"a", "b", "a", "b", "a", "b", "a", "b", "b"},
		[]string{"a", "a", "b"},
		[]string{"a", "a", "b", "a", "a", "b", "a", "a", "b", "a", "a", "b"},
		[]string{"one", "one", "two"},
	}
	for i, example := range examples {
		t.Run(fmt.Sprintf("example%d", i), func(t *testing.T) {
			result, err := ParseChain(example)
			if err != nil {
				t.Fatalf("Error parsing expression: %v", err)
			}
			if len(result) != len(expected[i]) {
				t.Fatalf("'%s' %v != %v", example, result, expected[i])
			}

		})
	}

}
