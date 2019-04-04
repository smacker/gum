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

	assert.True(t, zm.mappings.Has(src, dst.GetChild(0)))
	assert.True(t, zm.mappings.Has(src.GetChild(0), dst.GetChild(0).GetChild(0)))
	assert.True(t, zm.mappings.Has(src.GetChild(1), dst.GetChild(0).GetChild(1)))
	assert.True(t, zm.mappings.Has(src.GetChild(1).GetChild(0), dst.GetChild(0).GetChild(1).GetChild(0)))
	assert.True(t, zm.mappings.Has(src.GetChild(1).GetChild(2), dst.GetChild(0).GetChild(1).GetChild(2)))
}

func TestZsMatcherSlide(t *testing.T) {
	src, dst := readFixtures("testdata/zs/slide_src.json", "testdata/zs/slide_dst.json")

	zm := newZsMatcher()
	zm.Match(src, dst)

	assert.Equal(t, zm.mappings.Size(), 5)

	assert.True(t, zm.mappings.Has(src, dst))
	assert.True(t, zm.mappings.Has(src.GetChild(0).GetChild(0), dst.GetChild(0)))
	assert.True(t, zm.mappings.Has(src.GetChild(0).GetChild(0).GetChild(0), dst.GetChild(0).GetChild(0)))
	assert.True(t, zm.mappings.Has(src.GetChild(0).GetChild(1), dst.GetChild(1).GetChild(0)))
	assert.True(t, zm.mappings.Has(src.GetChild(0).GetChild(2), dst.GetChild(2)))
}
