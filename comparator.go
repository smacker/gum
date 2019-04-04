package gum

import (
	"math"
)

// mappingComparator provides Less(m1, m2 Mapping) function
// to sort mappings by the similarity value between nodes in each mapping
//
// check description of similarity function for details
type mappingComparator struct {
	maxTreeSize    int
	srcDescendants map[Tree]map[Tree]bool
	dstDescendants map[Tree]map[Tree]bool

	mappings     *mappingStore
	similarities map[Mapping]float64
}

// newMappingComparator creates mappingComparator for list of the list of ambiguous mappings
func newMappingComparator(ambiguousMappings []Mapping, mappings *mappingStore, maxTreeSize int) *mappingComparator {
	c := &mappingComparator{
		srcDescendants: make(map[Tree]map[Tree]bool),
		dstDescendants: make(map[Tree]map[Tree]bool),
		similarities:   make(map[Mapping]float64),

		mappings:    mappings,
		maxTreeSize: maxTreeSize,
	}

	for _, m := range ambiguousMappings {
		c.similarities[m] = c.similarity(m[0], m[1])
	}

	return c
}

// Less reports whether the mapping m1 should sort before the mapping m2
func (c *mappingComparator) Less(m1, m2 Mapping) bool {
	// mappings with greater similarities go first
	if c.similarities[m1] != c.similarities[m2] {
		return c.similarities[m1] > c.similarities[m2]
	}
	// mappings with left node closer to the root go first
	if m1[0].GetID() != m2[0].GetID() {
		return m1[0].GetID() < m2[0].GetID()
	}
	// mappings with right node closer to the root go first
	return m1[1].GetID() < m2[1].GetID()
}

// similarity return a value which indicates how similar src and dst Trees are.
// the value is calculated as a weighed sum of:
// - jaccard similarity of descendants for siblings
// - position in parent similarity
// - position in the tree (from root) similarity
func (c *mappingComparator) similarity(src, dst Tree) float64 {
	return 100*c.jaccardSimilarity(src.GetParent(), dst.GetParent()) +
		10*c.posInParentSimilarity(src, dst) + c.numberingSimilarity(src, dst)
}

// jaccard similarity of descendants
func (c *mappingComparator) jaccardSimilarity(src, dst Tree) float64 {
	num := float64(c.numberOfCommonDescendants(src, dst))
	den := float64(len(c.srcDescendants[src])+len(c.dstDescendants[dst])) - num
	return num / den
}

// descendants are common only if they appeared in the mapping
func (c *mappingComparator) numberOfCommonDescendants(src, dst Tree) int {
	if _, ok := c.srcDescendants[src]; !ok {
		for _, d := range getDescendants(src) {
			c.srcDescendants[src][d] = true
		}
	}
	if _, ok := c.dstDescendants[dst]; !ok {
		for _, d := range getDescendants(dst) {
			c.dstDescendants[dst][d] = true
		}
	}

	common := 0

	for t := range c.srcDescendants[src] {
		// skip nodes that didn't appear in the mapping
		m, ok := c.mappings.srcs[t]
		if !ok {
			continue
		}

		if _, ok := c.dstDescendants[dst][m]; ok {
			common++
		}
	}

	return common
}

func (c *mappingComparator) posInParentSimilarity(src, dst Tree) float64 {
	posSrc := 0
	maxSrcPos := 1
	if !isRoot(src) {
		posSrc = getChildPosition(src.GetParent(), src)
		maxSrcPos = len(src.GetParent().GetChildren())
	}
	posDst := 0
	maxDstPos := 1
	if !isRoot(dst) {
		posDst = getChildPosition(dst.GetParent(), dst)
		maxDstPos = len(dst.GetParent().GetChildren())
	}

	maxPosDiff := maxSrcPos
	if maxDstPos > maxPosDiff {
		maxPosDiff = maxDstPos
	}

	return 1 - (math.Abs(float64(posSrc)-float64(posDst)) / float64(maxPosDiff))
}

func (c *mappingComparator) numberingSimilarity(src, dst Tree) float64 {
	return 1 - math.Abs(float64(src.GetID())-float64(dst.GetID())/float64(c.maxTreeSize))
}
