package gum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZsMatcherSimple(t *testing.T) {
	src, dst := readFixtures("testdata/zs/src.json", "testdata/zs/dst.json")

	zm := newZsMatcher()
	zm.Match(src, dst)

	assert.Equal(t, zm.mappings.Size(), 5)

	assert.True(t, zm.mappings.Has(src, getChild(dst, 0)))
	assert.True(t, zm.mappings.Has(getChild(src, 0), getChild(dst, 0, 0)))
	assert.True(t, zm.mappings.Has(getChild(src, 1), getChild(dst, 0, 1)))
	assert.True(t, zm.mappings.Has(getChild(src, 1, 0), getChild(dst, 0, 1, 0)))
	assert.True(t, zm.mappings.Has(getChild(src, 1, 2), getChild(dst, 0, 1, 2)))
}

func TestZsMatcherSlide(t *testing.T) {
	src, dst := readFixtures("testdata/zs/slide_src.json", "testdata/zs/slide_dst.json")

	zm := newZsMatcher()
	zm.Match(src, dst)

	assert.Equal(t, zm.mappings.Size(), 5)

	assert.True(t, zm.mappings.Has(src, dst))
	assert.True(t, zm.mappings.Has(getChild(src, 0, 0), getChild(dst, 0)))
	assert.True(t, zm.mappings.Has(getChild(src, 0, 0, 0), getChild(dst, 0, 0)))
	assert.True(t, zm.mappings.Has(getChild(src, 0, 1), getChild(dst, 1, 0)))
	assert.True(t, zm.mappings.Has(getChild(src, 0, 2), getChild(dst, 2)))
}
