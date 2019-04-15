package gum

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// Tree is an internal representation of AST tree
type Tree struct {
	Type     string
	Value    string
	Children []*Tree

	id     int
	parent *Tree
	size   int
	height int
	hash   [16]byte
}

func (t *Tree) String() string {
	return fmt.Sprintf("%s%s%s", t.Type, "@@", t.Value)
}

func (t *Tree) isIsomorphicTo(o *Tree) bool {
	if o == nil {
		return false
	}

	return t.hash == o.hash
}

func (t *Tree) Refresh() {
	t.refresh(nil)
}

func (t *Tree) staticHashString() string {
	result := "[(" + t.String()
	for _, child := range t.Children {
		result += child.staticHashString()
	}
	return result + ")]"
}

func (t *Tree) isLeaf() bool {
	return len(t.Children) == 0
}

func (t *Tree) clone() *Tree {
	cl := *t

	children := make([]*Tree, len(t.Children))
	for i, c := range t.Children {
		children[i] = c.clone()
		children[i].parent = &cl
	}
	cl.Children = children

	return &cl
}

func (t *Tree) removeChild(child *Tree) {
	n := child
	newChildren := make([]*Tree, 0)
	for _, c := range t.Children {
		if c == n {
			continue
		}
		newChildren = append(newChildren, c)
	}
	t.Children = newChildren
}

func (t *Tree) addChild(pos int, child *Tree) {
	t.Children = append(t.Children[:pos], append([]*Tree{child}, t.Children[pos:]...)...)
}

func (t *Tree) refresh(parent *Tree) {
	for _, child := range t.Children {
		child.refresh(t)
	}

	if parent == nil {
		t.parent = nil
	} else {
		t.parent = parent
	}

	t.refreshSize()
	t.refreshHeight()
	t.refreshHash()
}

func (t *Tree) refreshSize() {
	for _, t := range postOrder(t) {
		n := t
		size := 1
		if !t.isLeaf() {
			for _, c := range t.Children {
				size += c.size
			}
		}
		n.size = size
	}
}

func (t *Tree) refreshHeight() {
	if t.isLeaf() {
		t.height = 1
		return
	}

	t.height = t.Children[0].height
	for _, child := range t.Children[1:] {
		if child.height > t.height {
			t.height = child.height
		}
	}
	t.height++
}

func (t *Tree) refreshHash() {
	t.hash = md5.Sum([]byte(t.staticHashString()))
}

func treeFromJSON(s string) (*Tree, error) {
	var t Tree
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return nil, err
	}
	t.refresh(nil)

	for i, v := range breadthFirst(&t) {
		v.id = i
	}

	return &t, nil
}

func isRoot(t *Tree) bool {
	return t.parent == nil
}

func preOrder(t *Tree) []*Tree {
	var trees []*Tree

	trees = append(trees, t)
	if !t.isLeaf() {
		for _, c := range t.Children {
			trees = append(trees, preOrder(c)...)
		}
	}

	return trees
}

func postOrder(t *Tree) []*Tree {
	var trees []*Tree

	if !t.isLeaf() {
		for _, c := range t.Children {
			trees = append(trees, postOrder(c)...)
		}
	}
	trees = append(trees, t)

	return trees
}

func breadthFirst(t *Tree) []*Tree {
	trees := make([]*Tree, 0)
	currents := []*Tree{t}
	for len(currents) > 0 {
		c := currents[0]
		currents = currents[1:]
		trees = append(trees, c)
		currents = append(currents, c.Children...)
	}

	return trees
}

func getDescendants(t *Tree) []*Tree {
	trees := preOrder(t)
	return trees[1:]
}

func getTrees(t *Tree) []*Tree {
	return preOrder(t)
}

func getChildPosition(t *Tree, child *Tree) int {
	idx := -1
	for i, c := range t.Children {
		if c == child {
			idx = i
			break
		}
	}
	return idx
}

func positionInParent(t *Tree) int {
	return getChildPosition(t.parent, t)
}
