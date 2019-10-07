package gum

import (
	"sort"
)

// subtreeMatcher implement top-down phase of GumTree algorithm
// greedy search of the greatest isomorphic subtrees
type subtreeMatcher struct {
	mappings *mappingStore
	// The algorithm considers only nodes with a height greater than MinHeight.
	// height of a node is:
	// - for a leaf node = 1
	// - for an internal node = max height of the children + 1
	MinHeight int
}

func newSubtreeMatcher() *subtreeMatcher {
	return &subtreeMatcher{MinHeight: defaultMinHeight, mappings: newMappingStore()}
}

// Match generates MappingStore with pair of nodes from src and dst Trees
func (m *subtreeMatcher) Match(src, dst *Tree) *mappingStore {
	maxTreeSize := src.size
	if dst.size > maxTreeSize {
		maxTreeSize = dst.size
	}

	mMapping := newMultiMapping()

	srcTrees := newPriorityTreeList(src, m.MinHeight)
	dstTrees := newPriorityTreeList(dst, m.MinHeight)

	// Map the common subtrees with the greatest height possible.
	// Start with the roots (since they have the greatest heights) and check if they are isomorphic.
	// If they are not, their children are then tested.
	// A node is matched as soon as an isomorphic node is found in the other tree.
	for srcTrees.PeekHeight() != -1 && dstTrees.PeekHeight() != -1 {
		// make tree lists the same height by removing tallest trees
		for srcTrees.PeekHeight() != dstTrees.PeekHeight() {
			m.popLarger(srcTrees, dstTrees)
		}

		currentHeightSrcTrees := srcTrees.Pop()
		currentHeightDstTrees := dstTrees.Pop()

		marksForSrcTrees := make([]bool, len(currentHeightSrcTrees))
		marksForDstTrees := make([]bool, len(currentHeightDstTrees))

		for i := 0; i < len(currentHeightSrcTrees); i++ {
			for j := 0; j < len(currentHeightDstTrees); j++ {
				src := currentHeightSrcTrees[i]
				dst := currentHeightDstTrees[j]

				if src.IsIsomorphicTo(dst) {
					mMapping.Link(src, dst)
					marksForSrcTrees[i] = true
					marksForDstTrees[j] = true
				}
			}
		}

		// add children of unmatched trees & repeat
		for i := 0; i < len(marksForSrcTrees); i++ {
			if marksForSrcTrees[i] == false {
				srcTrees.AddChildren(currentHeightSrcTrees[i])
			}
		}
		for j := 0; j < len(marksForDstTrees); j++ {
			if marksForDstTrees[j] == false {
				dstTrees.AddChildren(currentHeightDstTrees[j])
			}
		}

		srcTrees.UpdateHeight()
		dstTrees.UpdateHeight()
	}

	m.filterMappings(mMapping, maxTreeSize)

	return m.mappings
}

func (m *subtreeMatcher) filterMappings(mm *multiMapping, maxTreeSize int) {
	// When a given node can be matched to several nodes,
	// all the potential mappings are kept in a candidate mappings list.
	ambiguousList := make([]Mapping, 0)

	// map of already processed nodes
	ignored := make(map[*Tree]bool)

	for src := range mm.srcs {
		// for unique matches add nodes and all children to the mapping
		if mm.IsSrcUnique(src) {
			// FIXME ugly
			var dst *Tree
			for dst = range mm.srcs[src] {
				break
			}
			m.addMappingRecursively(src, dst)
			continue
		}

		// add the node to ambiguousList with all possible matches
		if _, ok := ignored[src]; !ok {
			dsts := mm.srcs[src]
			var dst *Tree
			for dst = range mm.srcs[src] {
				break
			}
			srcs := mm.dsts[dst]

			for src := range srcs {
				for dst := range dsts {
					ambiguousList = append(ambiguousList, Mapping{src, dst})
				}
			}

			for src := range srcs {
				ignored[src] = true
			}
		}
	}

	// rank the mappings by score
	comp := newMappingComparator(ambiguousList, m.mappings, maxTreeSize)
	sort.Slice(ambiguousList, func(i, j int) bool { return comp.Less(ambiguousList[i], ambiguousList[j]) })

	// Select the best ambiguous mappings
	srcIgnored := make(map[*Tree]bool)
	dstIgnored := make(map[*Tree]bool)
	m.retainBestMapping(ambiguousList, srcIgnored, dstIgnored)
}

func (m *subtreeMatcher) addMappingRecursively(src, dst *Tree) {
	srcTrees := getTrees(src)
	dstTrees := getTrees(dst)
	for i := 0; i < len(srcTrees); i++ {
		currentSrcTree := srcTrees[i]
		currentDstTree := dstTrees[i]
		m.addMapping(currentSrcTree, currentDstTree)
	}
}

func (m *subtreeMatcher) addMapping(src, dst *Tree) {
	m.mappings.Link(src, dst)
}

func (m *subtreeMatcher) popLarger(srcTrees, dstTrees *priorityTreeList) {
	if srcTrees.PeekHeight() > dstTrees.PeekHeight() {
		srcTrees.Open()
	} else {
		dstTrees.Open()
	}
}

func (m *subtreeMatcher) retainBestMapping(mappings []Mapping, srcIgnored, dstIgnored map[*Tree]bool) {
	for len(mappings) > 0 {
		mapping := mappings[0]
		mappings = mappings[1:]
		_, firstIgnored := srcIgnored[mapping[0]]
		_, secondIgnored := dstIgnored[mapping[1]]
		if !firstIgnored && !secondIgnored {
			m.addMappingRecursively(mapping[0], mapping[1])
			srcIgnored[mapping[0]] = true
			dstIgnored[mapping[1]] = true
		}
	}
}
