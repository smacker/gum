package gum

// Tree defines required methods on the AST nodes for gum diff
type Tree interface {
	GetID() int
	GetType() string
	GetLabel() string

	GetParent() Tree
	GetChildren() []Tree
	GetChild(int) Tree

	GetSize() int
	GetHeight() int

	IsLeaf() bool
	IsIsomorphicTo(Tree) bool

	Clone() Tree

	SetParent(Tree)
	RemoveChild(Tree)
	Refresh()

	String() string
}

func isRoot(t Tree) bool {
	return t.GetParent() == nil
}

func preOrder(t Tree) []Tree {
	var trees []Tree

	trees = append(trees, t)
	if !t.IsLeaf() {
		for _, c := range t.GetChildren() {
			trees = append(trees, preOrder(c)...)
		}
	}

	return trees
}

func postOrder(t Tree) []Tree {
	var trees []Tree

	if !t.IsLeaf() {
		for _, c := range t.GetChildren() {
			trees = append(trees, postOrder(c)...)
		}
	}
	trees = append(trees, t)

	return trees
}

func breadthFirst(t Tree) []Tree {
	trees := make([]Tree, 0)
	currents := []Tree{t}
	for len(currents) > 0 {
		c := currents[0]
		currents = currents[1:]
		trees = append(trees, c)
		currents = append(currents, c.GetChildren()...)
	}

	return trees
}

func getDescendants(t Tree) []Tree {
	trees := preOrder(t)
	return trees[1:]
}

func getTrees(t Tree) []Tree {
	return preOrder(t)
}

func getChildPosition(t Tree, child Tree) int {
	idx := -1
	for i, c := range t.GetChildren() {
		if c == child {
			idx = i
			break
		}
	}
	return idx
}

func positionInParent(t Tree) int {
	return getChildPosition(t.GetParent(), t)
}
