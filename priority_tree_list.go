package gum

// priorityTreeList is a height-indexed priority list
// this list contains a sequence of nodes, ordered by decreasing height
type priorityTreeList struct {
	trees      [][]Tree
	maxHeight  int
	currentIdx int

	minHeight int
}

func newPriorityTreeList(tree Tree, minHeight int) *priorityTreeList {
	ptl := &priorityTreeList{minHeight: minHeight}
	listSize := tree.GetHeight() - minHeight + 1
	if listSize < 0 {
		listSize = 0
	}
	if listSize == 0 {
		ptl.currentIdx = -1
	}

	ptl.trees = make([][]Tree, listSize)
	ptl.maxHeight = tree.GetHeight()
	ptl.addTree(tree)

	return ptl
}

// PeekHeight returns the greatest height of the list
func (ptl *priorityTreeList) PeekHeight() int {
	if ptl.currentIdx == -1 {
		return -1
	}
	return ptl.maxHeight - ptl.currentIdx
}

// Open returns and removes the list of all nodes with the largest height
// and adds all the children of that trees in the list
func (ptl *priorityTreeList) Open() []Tree {
	pop := ptl.Pop()
	if pop != nil {
		for _, tree := range pop {
			ptl.AddChildren(tree)
		}
		ptl.UpdateHeight()
		return pop
	}
	return nil
}

// AddChildren adds children of the Tree to the list
func (ptl *priorityTreeList) AddChildren(t Tree) {
	for _, c := range t.GetChildren() {
		ptl.addTree(c)
	}
}

// Pop returns and removes the list of all nodes with the largest height
func (ptl *priorityTreeList) Pop() []Tree {
	if ptl.currentIdx == -1 {
		return nil
	}

	pop := ptl.trees[ptl.currentIdx]
	ptl.trees[ptl.currentIdx] = nil
	return pop
}

// UpdateHeight re-calculates the height of the tree
func (ptl *priorityTreeList) UpdateHeight() {
	ptl.currentIdx = -1
	for i := 0; i < len(ptl.trees); i++ {
		if ptl.trees[i] != nil {
			ptl.currentIdx = i
			break
		}
	}
}

func (ptl *priorityTreeList) addTree(tree Tree) {
	if tree.GetHeight() < ptl.minHeight {
		return
	}

	idx := ptl.idx(tree)
	if ptl.trees[idx] == nil {
		ptl.trees[idx] = make([]Tree, 0)
	}
	ptl.trees[idx] = append(ptl.trees[idx], tree)
}

func (ptl *priorityTreeList) idx(tree Tree) int {
	return ptl.maxHeight - tree.GetHeight()
}
