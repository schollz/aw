package parser

import (
	"strconv"
	"strings"
)

func CalcP(s string) string {
	s = strings.ReplaceAll(s, "m", "96")
	s = strings.ReplaceAll(s, "h", "48")
	s = strings.ReplaceAll(s, "q", "24")
	s = strings.ReplaceAll(s, "s", "12")
	s = strings.ReplaceAll(s, "e", "6")
	return s
}

func HexToNum(s string) int {
	// Convert a one-letter hex (0-9, a-f) to a number
	if len(s) != 1 {
		return -1 // or handle error according to your needs
	}

	// Parse single hex digit to integer
	value, err := strconv.ParseInt(s, 16, 0)
	if err != nil {
		return -1 // or handle error appropriately
	}
	return int(value)
}

func min(values ...float64) float64 {
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}
