package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexToNum(t *testing.T) {
	assert.Equal(t, HexToNum("f"), 15)
}
