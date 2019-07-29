package gum

import (
	"fmt"
	"os"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionsSimple(t *testing.T) {
	// TODO: update fixture/mapping to include tree-insert action

	src, dst := readFixtures("testdata/actions/src.json", "testdata/actions/dst.json")

	mapping := make([]Mapping, 0)
	mapping = append(mapping, Mapping{src, dst})
	mapping = append(mapping, Mapping{getChild(src, 1), getChild(dst, 0)})
	mapping = append(mapping, Mapping{getChild(src, 1, 0), getChild(dst, 0, 0)})
	mapping = append(mapping, Mapping{getChild(src, 1, 0), getChild(dst, 0, 0)})
	mapping = append(mapping, Mapping{getChild(src, 1, 1), getChild(dst, 0, 1)})
	mapping = append(mapping, Mapping{getChild(src, 0), getChild(dst, 1, 0)})
	mapping = append(mapping, Mapping{getChild(src, 0, 0), getChild(dst, 1, 0, 0)})
	mapping = append(mapping, Mapping{getChild(src, 4), getChild(dst, 3)})
	mapping = append(mapping, Mapping{getChild(src, 4, 0), getChild(dst, 3, 0, 0, 0)})

	actions := Patch(src, dst, mapping)
	assert.Len(t, actions, 11)
	for _, a := range actions {
		fmt.Println(a)
	}

	apply(src, actions)

	a := actions[0]
	assert.Equal(t, Update, a.Type)
	assert.Equal(t, "0@@a", a.Node.String())
	assert.Equal(t, "z", a.Value)

	a = actions[1]
	assert.Equal(t, Insert, a.Type)
	assert.Equal(t, "0@@h", a.Node.String())
	assert.Equal(t, "0@@a", a.Parent.String())
	assert.Equal(t, 2, a.Pos)

	a = actions[2]
	assert.Equal(t, Insert, a.Type)
	assert.Equal(t, "0@@x", a.Node.String())
	assert.Equal(t, "0@@a", a.Parent.String())
	assert.Equal(t, 3, a.Pos)

	a = actions[3]
	assert.Equal(t, Move, a.Type)
	assert.Equal(t, "0@@e", a.Node.String())
	assert.Equal(t, "0@@h", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[4]
	assert.Equal(t, Insert, a.Type)
	assert.Equal(t, "0@@w", a.Node.String())
	assert.Equal(t, "0@@x", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[5]
	assert.Equal(t, Insert, a.Type)
	assert.Equal(t, "0@@u", a.Node.String())
	assert.Equal(t, "0@@j", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[6]
	assert.Equal(t, Update, a.Type)
	assert.Equal(t, "0@@f", a.Node.String())
	assert.Equal(t, "y", a.Value)

	a = actions[7]
	assert.Equal(t, a.Type, Insert)
	assert.Equal(t, "0@@v", a.Node.String())
	assert.Equal(t, "0@@u", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[8]
	assert.Equal(t, a.Type, Move)
	assert.Equal(t, "0@@k", a.Node.String())
	assert.Equal(t, "0@@v", a.Parent.String())
	assert.Equal(t, 0, a.Pos)

	a = actions[9]
	assert.Equal(t, a.Type, DeleteTree)
	assert.Equal(t, "0@@g", a.Node.String())

	a = actions[10]
	assert.Equal(t, a.Type, Delete)
	assert.Equal(t, "0@@i", a.Node.String())
}

func TestSimplify(t *testing.T) {
	tree1, _ := treeFromJSON(`{"root": {
		"typeLabel": "0",
		"label": "0",
		"children": [
			{
				"typeLabel": "1",
				"label": "1",
				"children": [
					{
						"typeLabel": "2",
						"label": "2",
						"children": []
					}
				]
			}
		]
	}}`)

	actions := []*Action{
		// next 3 actions should be replaced by tree-insert
		newInsert(tree1, nil, 0),
		newInsert(tree1.Children[0], nil, 0),
		newInsert(tree1.Children[0].Children[0], nil, 0),
		// no replace
		newUpdate(tree1, ""),
		newInsert(tree1, nil, 0),
		newUpdate(tree1, ""),
		newInsert(tree1, nil, 0),
		newInsert(tree1.Children[0], nil, 0),
		newDelete(tree1.Children[0]),
		newUpdate(tree1, ""),
		// next 3 actions should be replaced by tree-delete
		newDelete(tree1),
		newDelete(tree1.Children[0]),
		newDelete(tree1.Children[0].Children[0]),
	}

	g := &actionGenerator{}
	simplifiedActions := g.simplify(actions)

	assert.Len(t, simplifiedActions, 9)
}

// apply function uses ids, so actions must be generated from the same tree
// nb: changes trees inside actions
func apply(t *Tree, actions []*Action) *Tree {
	new := t.clone()
	nodes := getTrees(new)
	idToNode := make(map[int]*Tree, len(nodes))
	for _, n := range nodes {
		idToNode[n.GetID()] = n
	}

	for _, a := range actions {
		if a.Type == Insert {
			a.Node.id = 0 - a.Node.id
		}

		if a.Type == InsertTree {
			for _, t := range getTrees(a.Node) {
				t.id = 0 - t.id
			}
		}
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
			idToNode[a.Parent.GetID()].addChild(a.Pos, idToNode[nid])
		case Delete:
			idToNode[nid].GetParent().removeChild(idToNode[nid])
			delete(idToNode, nid)
		case DeleteTree:
			idToNode[nid].GetParent().removeChild(idToNode[nid])
			for _, d := range getDescendants(idToNode[nid]) {
				delete(idToNode, d.GetID())
			}
			delete(idToNode, nid)
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
	orgSrc, orgDst := readFixtures("testdata/paper/src.json", "testdata/paper/dst.json")
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
	deepCompare(t, orgDst, dst)

	fmt.Println("new tree")
	fmt.Println(treeString(changed))

	require.Equal(t, treeString(dst), treeString(changed))
}

func TestActionsApply2(t *testing.T) {
	src, dst := readFixtures("testdata/actions/src.json", "testdata/actions/dst.json")
	mappings := Match(src, dst)
	actions := Patch(src, dst, mappings)
	changed := apply(src, actions)

	require.Equal(t, treeString(dst), treeString(changed))
}

func TestActionsApply3(t *testing.T) {
	if _, err := os.Stat("testdata/parsed/samples"); os.IsNotExist(err) {
		t.Skip("directory with processed samples doesn't exist")
	}

	src, dst := readFixtures("testdata/parsed/samples/java/Example_v0.java", "testdata/parsed/samples/java/Example_v1.java")
	mappings := Match(src, dst)
	actions := Patch(src, dst, mappings)
	changed := apply(src, actions)

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
