package gum

const defaultMinHeight = 2
const defaultMaxSize = 100
const defaultSimThreshold = 0.5

// Mapping contains matched nodes from compared trees
type Mapping [2]Tree

// Action holds a tree node and action (insert/delete/update/move)
type Action interface {
	Type() string
	Node() Tree
}

// Matcher implements GumTree algorithm to compare abstract syntax trees
type Matcher struct {
	// MinHeight limits nodes considered by top-down phase
	// recommended MinHeight = 2 to avoid single identifiers to match everywhere
	MinHeight int
	// MaxSize is used in the recovery part of bottom-up phase that can trigger a cubic algorithm
	// recommended MaxSize = 100 to avoid long computation times
	MaxSize int
	// SimThreshold minimum ratio for common descendants between two nodes given a set of mappings
	// recommended SimThreshold = 0.5 because
	// under 50% of common nodes, two container nodes are probably different
	SimThreshold float64
}

// NewMatcher creates new Matcher with default (recommended) parameters
func NewMatcher() *Matcher {
	return &Matcher{
		MinHeight:    defaultMinHeight,
		MaxSize:      defaultMaxSize,
		SimThreshold: defaultSimThreshold,
	}
}

// Match generate list on mappings (pairs of nodes) that are considered similar in both trees
func (m *Matcher) Match(src, dst Tree) []Mapping {
	sm := newSubtreeMatcher()
	sm.MinHeight = m.MinHeight
	mappings := sm.Match(src, dst)

	bum := newBottomUpMatcher(mappings)
	bum.maxSize = m.MaxSize
	bum.simThreshold = m.SimThreshold
	return bum.Match(src, dst).ToList()
}
