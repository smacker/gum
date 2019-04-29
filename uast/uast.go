package uast

import (
	"sort"

	"github.com/smacker/gum"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

// ToTree converts bblfsh.Node to gum.Tree
func ToTree(n nodes.Node) *gum.Tree {
	t := toTree(n)
	t.Refresh()
	return t
}

func toTree(n nodes.Node) *gum.Tree {
	children := getChildren(n)
	tree := &gum.Tree{
		Type:     uast.TypeOf(n),
		Value:    uast.TokenOf(n),
		Children: make([]*gum.Tree, len(children)),
	}
	for i, child := range children {
		tree.Children[i] = toTree(child)
	}

	return tree
}

func getChildren(n nodes.Node) []nodes.Node {
	var children []nodes.Node
	switch n := n.(type) {
	case nodes.Array:
		children = n
	case nodes.Object:
		var i int
		keys := make([]string, len(n))
		for k := range n {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		for _, k := range keys {
			if k == uast.KeyRoles {
				continue
			}
			v := n[k]
			switch v := v.(type) {
			case nodes.Object:
				children = append(children, v)
			case nodes.Array:
				children = append(children, v...)
			}
		}
	}

	var filtered []nodes.Node
	for _, child := range children {
		if uast.TypeOf(child) != uast.TypePositions {
			filtered = append(filtered, child)
		}
	}

	return filtered
}
