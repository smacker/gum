package gum

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

// FIXME
// func TestMinHeightThreshold(t *testing.T) {
// 	src, dst := readFixtures("testdata/gumtree/src.json", "testdata/gumtree/dst.json")

// 	m := NewMatcher()
// 	m.MinHeight = 0
// 	m.MaxSize = 0
// 	mappings := m.Match(src, dst)

// 	assert.Len(t, mappings, 5)

// 	m = NewMatcher()
// 	m.MinHeight = 1
// 	m.MaxSize = 0
// 	mappings = m.Match(src, dst)

// 	assert.Len(t, mappings, 4)
// }

// func TestMinSizeThreshold(t *testing.T) {
// 	src, dst := readFixtures("testdata/gumtree/src.json", "testdata/gumtree/dst.json")

// 	m := NewMatcher()
// 	m.MinHeight = 0
// 	m.MaxSize = 5
// 	mappings := m.Match(src, dst)

// 	assert.Len(t, mappings, 6)
// }

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
	fmt.Println(strings.Repeat("-", tab), t)
	for _, c := range t.Children {
		treePrint(c, tab+1)
	}
}

func TestActionsTODO(t *testing.T) {
	src, dst := readFixtures("testdata/actions/src.json", "testdata/actions/dst.json")

	mapping := make([]Mapping, 0)
	mapping = append(mapping, Mapping{src, dst})
	mapping = append(mapping, Mapping{getChild(src, 1), getChild(dst, 0)})
	mapping = append(mapping, Mapping{getChild(src, 1, 0), getChild(dst, 0, 0)})
	mapping = append(mapping, Mapping{getChild(src, 1, 1), getChild(dst, 0, 1)})
	mapping = append(mapping, Mapping{getChild(src, 0), getChild(dst, 1, 0)})
	mapping = append(mapping, Mapping{getChild(src, 0, 0), getChild(dst, 1, 0, 0)})
	mapping = append(mapping, Mapping{getChild(src, 4), getChild(dst, 3)})
	mapping = append(mapping, Mapping{getChild(src, 4, 0), getChild(dst, 3, 0, 0, 0)})

	actions := Patch(src, dst, mapping)
	assert.Len(t, actions, 9)

	a := actions[0]
	assert.Equal(t, a.Type, Insert)
	assert.Equal(t, "0@@h", a.Node.String())
	assert.Equal(t, "0@@a", a.Parent.String())
	assert.Equal(t, 2, a.Pos)

	a = actions[1]
	assert.Equal(t, a.Type, InsertTree)
	assert.Equal(t, "0@@x", a.Node.String())
	assert.Equal(t, "0@@a", a.Parent.String())
	assert.Equal(t, 3, a.Pos)

	a = actions[2]
	assert.Equal(t, a.Type, Move)
	assert.Equal(t, "0@@e", a.Node.String())
	assert.Equal(t, "0@@h", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[3]
	assert.Equal(t, a.Type, Insert)
	assert.Equal(t, "0@@u", a.Node.String())
	assert.Equal(t, "0@@j", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[4]
	assert.Equal(t, a.Type, Update)
	assert.Equal(t, "0@@f", a.Node.String())
	assert.Equal(t, "y", a.Value)

	a = actions[5]
	assert.Equal(t, a.Type, Insert)
	assert.Equal(t, "0@@v", a.Node.String())
	assert.Equal(t, "0@@u", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[6]
	assert.Equal(t, a.Type, Move)
	assert.Equal(t, "0@@k", a.Node.String())
	assert.Equal(t, "0@@v", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[7]
	assert.Equal(t, a.Type, DeleteTree)
	assert.Equal(t, "0@@g", a.Node.String())

	a = actions[8]
	assert.Equal(t, a.Type, Delete)
	assert.Equal(t, "0@@i", a.Node.String())
}

func getChild(t *Tree, path ...int) *Tree {
	for _, i := range path {
		t = t.Children[i]
	}
	return t
}
