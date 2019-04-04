package gum

import (
	"math"
)

// zsMatcher implements ZhangShasha algorithm
//
// Original paper (1989):
// http://www.grantjenks.com/wiki/_media/ideas:simple_fast_algorithms_for_the_editing_distance_between_tree_and_related_problems.pdf
// Simpler explanation of the algorithm:
// http://research.cs.queensu.ca/TechReports/Reports/1995-372.pdf
type zsMatcher struct {
	zsSrc *zsTree
	zsDst *zsTree

	// distance between 2 nodes considered separately from sibling & ancestors (but not descendants)
	treeDist [][]float64

	// distance between 2 nodes considered in the context of their left siblings
	forestDist [][]float64
	// example:
	//        T                T'
	//    /      \          /     \
	//   a        b        q       b
	//  / \     / | \     / \    / | \
	// c   d   e  f  g   r   s  t  u  v
	//
	// treeDist(b in T, b in T') == 3
	// forestDist(b in T, b in T') == 6
	// treeDist(a) == forestDist(q) == 3
	// if each change operation == 1

	mappings *mappingStore
}

func newZsMatcher() *zsMatcher {
	return &zsMatcher{mappings: newMappingStore()}
}

func (m *zsMatcher) Match(src, dst Tree) {
	m.zsSrc = newZsTree(src)
	m.zsDst = newZsTree(dst)

	// compute forest distance matrix for keyroots
	m.computeTreeDist(src, dst)

	rootNodePair := true
	treePairs := make([][]int, 0)
	// start from the roots
	treePairs = append([][]int{{m.zsSrc.nodeCount, m.zsDst.nodeCount}}, treePairs...)
	for len(treePairs) > 0 {
		var treePair []int
		treePair, treePairs = treePairs[0], treePairs[1:]
		lastRow := treePair[0]
		lastCol := treePair[1]
		// compute forest distance matrix
		if !rootNodePair {
			m.fillForestDist(lastRow, lastCol)
		}

		rootNodePair = false
		// compute mapping for current forest distance matrix
		firstRow := m.zsSrc.lld(lastRow) - 1
		firstCol := m.zsDst.lld(lastCol) - 1
		row := lastRow
		col := lastCol

		for (row > firstRow) || (col > firstCol) {
			if (row > firstRow) && (m.forestDist[row-1][col]+1 == m.forestDist[row][col]) {
				// node with postorder id = row is deleted from t1
				row--
			} else if (col > firstCol) && (m.forestDist[row][col-1]+1 == m.forestDist[row][col]) {
				// node with postorder id = col is inserted into t2
				col--
			} else {
				// node with postorder id = row in t1 is renamed to node col in t2
				if (m.zsSrc.lld(row)-1 == m.zsSrc.lld(lastRow)-1) && (m.zsDst.lld(col)-1 == m.zsDst.lld(lastCol)-1) {
					// if both subforests are trees, map nodes
					tSrc := m.zsSrc.tree(row)
					tDst := m.zsDst.tree(col)
					if tSrc.GetType() == tDst.GetType() {
						m.addMapping(tSrc, tDst)
					} else {
						panic("Should not map incompatible nodes.")
					}
					row--
					col--
				} else {
					// pop subtree pair
					treePairs = append([][]int{{row, col}}, treePairs...)
					// continue with forest to the left of the popped
					// subtree pair

					row = m.zsSrc.lld(row) - 1
					col = m.zsDst.lld(col) - 1
				}
			}
		}
	}
}

// computeTreeDist calculates tree and forest distances for keyroots
func (m *zsMatcher) computeTreeDist(src, dst Tree) {
	srcSize := src.GetSize() + 1
	dstSize := dst.GetSize() + 1

	m.treeDist = make([][]float64, srcSize)
	m.forestDist = make([][]float64, srcSize)
	for i := 0; i < srcSize; i++ {
		m.treeDist[i] = make([]float64, dstSize)
		m.forestDist[i] = make([]float64, dstSize)
	}

	for i := 1; i < len(m.zsSrc.kr); i++ {
		for j := 1; j < len(m.zsDst.kr); j++ {
			m.fillForestDist(m.zsSrc.kr[i], m.zsDst.kr[j])
		}
	}
}

func (m *zsMatcher) fillForestDist(i, j int) {
	zsSrc := m.zsSrc
	zsDst := m.zsDst

	// permanent matrix
	treeDist := m.treeDist
	// temporary matrix
	forestDist := m.forestDist

	// fill the part of forest dist matrix
	// from left-most children of i & j to i & j
	forestDist[zsSrc.lld(i)-1][zsDst.lld(j)-1] = 0 // forestDist(0, 0)
	// go up from left-most leaf to the keyroot
	for di := zsSrc.lld(i); di <= i; di++ {
		costDel := m.getDeletionCost(zsSrc.tree(di))
		forestDist[di][zsDst.lld(j)-1] = forestDist[di-1][zsDst.lld(j)-1] + costDel // forestDist(0, j)
		for dj := zsDst.lld(j); dj <= j; dj++ {
			costIns := m.getInsertionCost(zsDst.tree(dj))
			forestDist[zsSrc.lld(i)-1][dj] = forestDist[zsSrc.lld(i)-1][dj-1] + costIns // foredist(i, 0)

			if zsSrc.lld(di) == zsSrc.lld(i) && (zsDst.lld(dj) == zsDst.lld(j)) {
				costUpd := m.getUpdateCost(zsSrc.tree(di), zsDst.tree(dj))
				forestDist[di][dj] = math.Min(math.Min(forestDist[di-1][dj]+costDel,
					forestDist[di][dj-1]+costIns),
					forestDist[di-1][dj-1]+costUpd)
				treeDist[di][dj] = forestDist[di][dj]
			} else {
				forestDist[di][dj] = math.Min(math.Min(forestDist[di-1][dj]+costDel,
					forestDist[di][dj-1]+costIns),
					forestDist[zsSrc.lld(di)-1][zsDst.lld(dj)-1]+treeDist[di][dj])
			}
		}
	}
}

func (m *zsMatcher) getDeletionCost(t Tree) float64 {
	return 1
}

func (m *zsMatcher) getInsertionCost(t Tree) float64 {
	return 1
}

func (m *zsMatcher) getUpdateCost(n1, n2 Tree) float64 {
	if n1.GetType() == n2.GetType() {
		if n1.GetLabel() == "" || n2.GetLabel() == "" {
			return 1
		}

		return 1 - qGramsDistance().Compare(n1.GetLabel(), n2.GetLabel())
	}

	return math.MaxFloat64
}

func (m *zsMatcher) addMapping(src, dst Tree) {
	m.mappings.Link(src, dst)
}

type zsTree struct {
	nodeCount int
	leafCount int

	// llds[i] stores the postorder-id of the left-most leaf descendant of the i-th node in postorder
	llds []int
	// labels[i] is the tree of the i-th node in postorder
	labels []Tree

	// keyroots is the root of the tree + all nodes with a left sibling
	kr []int
}

func newZsTree(t Tree) *zsTree {
	nodeCount := t.GetSize()
	zt := &zsTree{
		nodeCount: nodeCount,
		llds:      make([]int, nodeCount),
		labels:    make([]Tree, nodeCount),
	}

	// fill labels and llds
	tmpData := make(map[Tree]int)
	for i, n := range postOrder(t) {
		tmpData[n] = i
		zt.labels[i] = n
		zt.llds[i] = tmpData[getFirstLeaf(n)]

		if n.IsLeaf() {
			zt.leafCount++
		}
	}

	// calculate keyroots
	zt.kr = make([]int, zt.leafCount+1)
	visited := make([]bool, zt.nodeCount+1)

	// example tree: postorder-id(left-most leaf id)
	//          7(1)
	//        /      \
	//    3(1)        6(4)
	//   /    \      /   \
	// 1(1)  2(2)  4(4)  5(5)
	//
	// krs: 2, 5, 6, 7
	k := len(zt.kr) - 1
	for i := zt.nodeCount; i >= 1; i-- {
		if !visited[zt.lld(i)] {
			zt.kr[k] = i
			visited[zt.lld(i)] = true
			k--
		}
	}

	return zt
}

// lld returns postorder-id of the left-most leaf descendant of the i-th node in postorder
func (t *zsTree) lld(i int) int {
	return t.llds[i-1] + 1
}

// tree returns the tree of the i-th node in postorder
func (t *zsTree) tree(i int) Tree {
	return t.labels[i-1]
}

func getFirstLeaf(t Tree) Tree {
	current := t
	for !current.IsLeaf() {
		current = current.GetChild(0)
	}

	return current
}
