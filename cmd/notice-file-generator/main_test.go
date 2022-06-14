package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOf(t *testing.T) {
	deps := []string{"a", "b"}
	assert.Equal(t, -1, IndexOf(deps, "c"))
	assert.Equal(t, 0, IndexOf(deps, "a"))
	assert.Equal(t, 1, IndexOf(deps, "b"))
}

func TestGenerateFileName(t *testing.T) {
	fileName := GenerateFileName("Sample File Name / Owww[]\"")

	assert.Equal(t, "Sample-File-Name-Owww-", fileName)
}
