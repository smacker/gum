package gum

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type node struct {
	Type     string
	Value    string
	Children []*node

	id     int
	parent Tree
	size   int
	height int
	hash   [16]byte
}

func (t *node) GetType() string {
	return t.Type
}

func (t *node) GetLabel() string {
	return t.Value
}

func (t *node) GetID() int {
	return t.id
}

func (t *node) GetParent() Tree {
	if t.parent == nil {
		return nil
	}
	return t.parent
}

func (t *node) GetSize() int {
	return t.size
}

func (t *node) GetHeight() int {
	return t.height
}

func (t *node) GetChildren() []Tree {
	result := make([]Tree, len(t.Children))
	for i, child := range t.Children {
		result[i] = child
	}
	return result
}

func (t *node) GetChild(i int) Tree {
	return t.Children[i]
}

func (t *node) IsIsomorphicTo(o Tree) bool {
	if o == nil {
		return false
	}

	return t.hash == o.(*node).hash
}

func (t *node) String() string {
	return fmt.Sprintf("%s%s%s", t.Type, "@@", t.Value)
}

func (t *node) StaticHashString() string {
	result := "[(" + t.String()
	for _, child := range t.Children {
		result += child.StaticHashString()
	}
	return result + ")]"
}

func (t *node) IsLeaf() bool {
	return len(t.Children) == 0
}

func (t *node) Clone() Tree {
	cl := *t

	children := make([]*node, len(t.Children))
	for i, c := range t.Children {
		children[i] = c.Clone().(*node)
	}
	cl.Children = children
	//cl.parent = t.parent.Clone().(*node)

	return &cl
}

func (t *node) SetParent(p Tree) {
	// if p != nil {
	// 	t.parent = p.(*node)
	// } else {
	// 	t.parent = nil
	// }
	t.parent = p
}

func (t *node) RemoveChild(child Tree) {
	n := child.(*node)
	newChildren := make([]*node, 0)
	for _, c := range t.Children {
		if c == n {
			continue
		}
		newChildren = append(newChildren, c)
	}
}

func (t *node) Refresh() {
	t.refresh(nil)
}

func (t *node) refresh(parent *node) {
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

func (t *node) refreshSize() {
	for _, t := range postOrder(t) {
		n := t.(*node)
		size := 1
		if !t.IsLeaf() {
			for _, c := range t.GetChildren() {
				size += c.GetSize()
			}
		}
		n.size = size
	}
}

func (t *node) refreshHeight() {
	if t.IsLeaf() {
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

func (t *node) refreshHash() {
	t.hash = md5.Sum([]byte(t.StaticHashString()))
}

func nodeFromJSON(s string) (*node, error) {
	var t node
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return nil, err
	}
	t.refresh(nil)

	for i, v := range breadthFirst(&t) {
		v.(*node).id = i
	}

	return &t, nil
}
