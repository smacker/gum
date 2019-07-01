package gum

import (
	"fmt"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionsSimple(t *testing.T) {
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

// FIXME: this function isn't really correct because it relies on ids of action.Node
// for "insert", action contains real node from dst tree with id from dst tree
// it can overlap with existing ids in source tree and everything will blow up
func apply(t *Tree, actions []*Action) *Tree {
	new := t.clone()
	nodes := getTrees(new)
	idToNode := make(map[int]*Tree, len(nodes))
	for _, n := range nodes {
		idToNode[n.GetID()] = n
	}

	for _, a := range actions {
		before := treeString(new)
		fmt.Println(a, "--->")

		n := a.Node.clone()
		nid := n.GetID()

		switch a.Type {
		case Insert:
			n.Children = nil
			idToNode[a.Parent.GetID()].addChild(a.Pos, n)
			idToNode[nid] = n
		case InsertTree:
			idToNode[a.Parent.GetID()].addChild(a.Pos, n)
			idToNode[nid] = n
		case Update:
			idToNode[nid].Value = a.Value
		case Move:
			idToNode[nid].GetParent().removeChild(idToNode[nid])
			idToNode[a.Parent.GetID()].addChild(a.Pos, n)
		case Delete:
			idToNode[nid].GetParent().removeChild(idToNode[nid])
			delete(idToNode, nid)
		case DeleteTree:
			idToNode[nid].GetParent().removeChild(idToNode[nid])
			delete(idToNode, nid)
			for _, d := range getDescendants(idToNode[nid]) {
				delete(idToNode, d.GetID())
			}
		default:
			panic("unsupported")
		}

		after := treeString(new)
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(before, after, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		fmt.Println("")
	}

	return new
}

func TestActionsApply(t *testing.T) {
	orgSrc, _ := readFixtures("testdata/paper/src.json", "testdata/paper/dst.json")
	src, dst := readFixtures("testdata/paper/src.json", "testdata/paper/dst.json")
	mappings := Match(src, dst)
	actions := Patch(src, dst, mappings)

	// fmt.Println("src tree:")
	// treePrint(src, 0)
	// fmt.Println("dst tree:")
	// treePrint(dst, 0)

	changed := apply(src, actions)

	// to make sure apply function didn't mess up anything
	deepCompare(t, orgSrc, src)

	fmt.Println("new tree")
	fmt.Println(treeString(changed))

	require.Equal(t, treeString(dst), treeString(changed))
}

func treeString(t *Tree) string {
	result := "(" + nodeString(t)
	for _, child := range t.Children {
		result += " " + treeString(child)
	}
	result += ")"
	return result
}

func nodeString(t *Tree) string {
	if t.Value != "" {
		return fmt.Sprintf("%s[%s]", t.Type, t.Value)
	}
	return t.Type
}
