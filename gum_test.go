package gum

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uses fixtures created from examples in the paper
func TestPaperValidation(t *testing.T) {
	src, dst := readFixtures("testdata/paper/src.json", "testdata/paper/dst.json")

	sm := newSubtreeMatcher()
	mappings := sm.Match(src, dst)

	assert.Equal(t, 10, mappings.Size())

	s := getChild(src, 0, 2, 1)
	d := getChild(dst, 0, 2, 1)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 3)
	d = getChild(dst, 0, 2, 3)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 4, 0, 0)
	d = getChild(dst, 0, 2, 4, 0, 0)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 4, 0, 1)
	d = getChild(dst, 0, 2, 4, 0, 2, 1)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)

	bum := newBottomUpMatcher(mappings)
	bum.simThreshold = 0.2
	mappings = bum.Match(src, dst)

	// 15 = 10 from top-down + 5 containers + 4 recovery mapping
	assert.Equal(t, 19, mappings.Size())

	// containers
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", src, dst)
	s = getChild(src, 0)
	d = getChild(dst, 0)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2)
	d = getChild(dst, 0, 2)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 4)
	d = getChild(dst, 0, 2, 4)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 4, 0)
	d = getChild(dst, 0, 2, 4, 0)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)

	// recovery
	s = getChild(src, 0, 0)
	d = getChild(dst, 0, 0)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 1)
	d = getChild(dst, 0, 1)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 0)
	d = getChild(dst, 0, 2, 0)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
	s = getChild(src, 0, 2, 2)
	d = getChild(dst, 0, 2, 2)
	assert.True(t, mappings.Has(s, d), "%v = %v mapping not found", s, d)
}

func TestMinHeightThreshold(t *testing.T) {
	// FIXME
	t.Skip("it doesn't work")

	src, dst := readFixtures("testdata/gumtree/src.json", "testdata/gumtree/dst.json")

	m := NewMatcher()
	m.MinHeight = 0
	m.MaxSize = 0
	mappings := m.Match(src, dst)

	assert.Len(t, mappings, 5)

	m = NewMatcher()
	m.MinHeight = 1
	m.MaxSize = 0
	mappings = m.Match(src, dst)

	assert.Len(t, mappings, 4)
}

func TestMinSizeThreshold(t *testing.T) {
	src, dst := readFixtures("testdata/gumtree/src.json", "testdata/gumtree/dst.json")

	m := NewMatcher()
	m.MinHeight = 0
	m.MaxSize = 5
	mappings := m.Match(src, dst)

	assert.Len(t, mappings, 6)
}

func readFixtures(fSrc, fDst string) (*Tree, *Tree) {
	srcJSON, err := ioutil.ReadFile(fSrc)
	if err != nil {
		panic(err)
	}
	dstJSON, err := ioutil.ReadFile(fDst)
	if err != nil {
		panic(err)
	}

	src, err := treeFromJSON(string(srcJSON))
	if err != nil {
		panic(err)
	}
	dst, err := treeFromJSON(string(dstJSON))
	if err != nil {
		panic(err)
	}

	return src, dst
}

func treePrint(t *Tree, tab int) {
	fmt.Println(strings.Repeat("-", tab), t, t.GetID())
	for _, c := range t.Children {
		treePrint(c, tab+1)
	}
}

func getChild(t *Tree, path ...int) *Tree {
	for _, i := range path {
		t = t.Children[i]
	}
	return t
}

func TestNotModified(t *testing.T) {
	orgSrc, orgDst := readFixtures("testdata/paper/src.json", "testdata/paper/dst.json")

	src, dst := readFixtures("testdata/paper/src.json", "testdata/paper/dst.json")
	mappings := Match(src, dst)
	deepCompare(t, orgSrc, src)
	deepCompare(t, orgDst, dst)

	Patch(src, dst, mappings)
	deepCompare(t, orgSrc, src)
	deepCompare(t, orgDst, dst)
}

func deepCompare(t *testing.T, a *Tree, b *Tree) {
	aTrees := getTrees(a)
	bTrees := getTrees(b)
	require.Equal(t, len(bTrees), len(aTrees))

	for i, bT := range bTrees {
		aT := aTrees[i]

		require.Equal(t, bT.size, aT.size)
		require.Equal(t, bT.height, aT.height)
		require.Equal(t, b.id, a.id)
		require.Equal(t, b.Type, a.Type)
		require.Equal(t, b.Value, a.Value)
		require.Equal(t, len(b.Children), len(a.Children))
	}
}
