package gum

// bottomUpMatcher implement bottom-up phase of GumTree algorithm
//
// it looks for container mappings first
// that are established when two nodes have a significant number of matching descendants
//
// for each container mapping found, it looks for recovery mappings,
// that are searched among the still un-matched descendants of the mapping's nodes
type bottomUpMatcher struct {
	mappings     *mappingStore
	maxSize      int
	simThreshold float64

	mappedSrc map[int]*Tree
	mappedDst map[int]*Tree

	srcIds map[int]*Tree
	dstIds map[int]*Tree
}

// newBottomUpMatcher requires mappings input from previous phase
func newBottomUpMatcher(mappings *mappingStore) *bottomUpMatcher {
	mappedSrc := make(map[int]*Tree, len(mappings.srcs))
	mappedDst := make(map[int]*Tree, len(mappings.srcs))
	for left, right := range mappings.srcs {
		putTrees(mappedSrc, left)
		putTrees(mappedDst, right)
	}

	return &bottomUpMatcher{
		mappings:     mappings,
		maxSize:      defaultMaxSize,
		simThreshold: defaultSimThreshold,
		mappedSrc:    mappedSrc,
		mappedDst:    mappedDst,
	}
}

// Match generates MappingStore with pair of nodes from src and dst trees
// taking into account previously mapped nodes
func (m *bottomUpMatcher) Match(src, dst *Tree) *mappingStore {
	m.srcIds = make(map[int]*Tree, src.size)
	m.dstIds = make(map[int]*Tree, dst.size)
	putTrees(m.srcIds, src)
	putTrees(m.dstIds, dst)

	for _, t := range PostOrder(src) {
		// when reach the root of the src tree
		// always map roots (cause they are "program" nodes)
		// and stop
		if isRoot(t) {
			m.addMapping(t, dst)
			m.lastChanceMatch(t, dst)
			break
		}

		// this algorithm ignores already matched nodes and leafs
		if !m.isSrcMatched(t) && !t.isLeaf() {
			candidates := m.getDstCandidates(t)

			// get the best candidate using jaccard similarity of descendants
			// limited by similarity threshold
			var best *Tree
			max := float64(-1)
			for _, cand := range candidates {
				sim := m.jaccardSimilarity(t, cand)
				if sim > max && sim >= m.simThreshold {
					max = sim
					best = cand
				}
			}

			if best != nil {
				m.lastChanceMatch(t, best)
				m.addMapping(t, best)
			}
		}
	}

	return m.mappings
}

// recovery mappings
//
// apply Zhang Shasha algorithm
// for descendants of container nodes without previously matched nodes
// if any of result trees have a size smaller than maxSize
func (m *bottomUpMatcher) lastChanceMatch(src, dst *Tree) {
	cSrc := src.clone()
	cDst := dst.clone()

	m.removeMatched(cSrc, true)
	m.removeMatched(cDst, false)

	// I follow reference implementation here
	// in the paper algorithm applied only if both resulting subtrees have a size smaller than maxSize
	// TODO: investigate how it affects accuracy, it's dangerous in terms of computation time
	if cSrc.size < m.maxSize || cDst.size < m.maxSize {
		zsm := newZsMatcher()
		zsm.Match(cSrc, cDst)
		for lt, rt := range zsm.mappings.srcs {
			left := m.srcIds[lt.id]
			right := m.dstIds[rt.id]

			if left.id == src.id || right.id == dst.id {
				//fmt.Printf("Trying to map already mapped source node (%v == %v || %v == %v)\n", left, src, right, dst)
				continue
			} else if !m.isMappingAllowed(left, right) {
				//fmt.Printf("Trying to map incompatible nodes (%v, %v)\n", left, right)
				continue
			} else if left.parent.Type != right.parent.Type {
				//fmt.Printf("Trying to map nodes with incompatible parents (%v, %v)\n", left.parent, right.parent)
				continue
			} else {
				m.addMapping(left, right)
			}
		}
	}

	putTrees(m.mappedSrc, src)
	putTrees(m.mappedDst, dst)
}

func (m *bottomUpMatcher) isMappingAllowed(src, dst *Tree) bool {
	return src.Type == dst.Type && !(m.isSrcMatched(src) || m.isDstMatched(dst))
}

// creates new subtree without previously matched nodes
func (m *bottomUpMatcher) removeMatched(tree *Tree, isSrc bool) *Tree {
	for _, t := range getTrees(tree) {
		if (isSrc && m.isSrcMatched(t)) || ((!isSrc) && m.isDstMatched(t)) {
			if t.parent != nil {
				t.parent.removeChild(t)
			}
			t.parent = nil
		}
	}
	tree.refresh(nil)

	return tree
}

// dst node is a candidate if:
// - it's unmatched yet
// - label is equal to src node
// - source and dst node have some matching descendants
func (m *bottomUpMatcher) getDstCandidates(src *Tree) []*Tree {
	// list dst descendants nodes that were matched previously
	seeds := make([]*Tree, 0)
	for _, c := range getDescendants(src) {
		if mp, ok := m.mappings.srcs[c]; ok {
			seeds = append(seeds, mp)
		}
	}

	// any parents of seeds if they match requirements
	candidates := make([]*Tree, 0)
	visited := make(map[*Tree]bool)
	for _, seed := range seeds {
		for {
			p := seed.parent
			if p == nil { // skip root nodes, they are special case
				break
			}
			if _, ok := visited[p]; ok {
				break
			}
			visited[p] = true

			if p.Type == src.Type && !m.isDstMatched(p) && !isRoot(p) {
				candidates = append(candidates, p)
			}

			seed = p
		}
	}

	return candidates
}

func (m *bottomUpMatcher) isSrcMatched(t *Tree) bool {
	_, ok := m.mappedSrc[t.id]
	return ok
}

func (m *bottomUpMatcher) isDstMatched(t *Tree) bool {
	_, ok := m.mappedDst[t.id]
	return ok
}

// jaccard similarity of mapped descendants
func (m *bottomUpMatcher) jaccardSimilarity(src, dst *Tree) float64 {
	srcDesc := getDescendants(src)
	dstDesc := getDescendants(dst)

	num := m.numberOfCommonDescendants(srcDesc, dstDesc)
	den := len(srcDesc) + len(dstDesc) - num
	return float64(num) / float64(den)
}

func (m *bottomUpMatcher) numberOfCommonDescendants(srcDesc, dstDesc []*Tree) int {
	dstDescendants := make(map[*Tree]bool, len(dstDesc))
	for _, t := range dstDesc {
		dstDescendants[t] = true
	}

	common := 0
	for _, t := range srcDesc {
		m, ok := m.mappings.srcs[t]
		if !ok {
			continue
		}

		if _, ok := dstDescendants[m]; ok {
			common++
		}

	}

	return common
}

func (m *bottomUpMatcher) addMapping(src, dst *Tree) {
	m.mappedSrc[src.id] = src
	m.mappedDst[dst.id] = dst
	m.mappings.Link(src, dst)
}
