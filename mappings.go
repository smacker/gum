package gum

// mappingStore holds results of the mapping
type mappingStore struct {
	srcs map[*Tree]*Tree
	dsts map[*Tree]*Tree
}

// newMappingStore creates new MappingStore
func newMappingStore() *mappingStore {
	return &mappingStore{
		srcs: make(map[*Tree]*Tree),
		dsts: make(map[*Tree]*Tree),
	}
}

// Link adds mapping to the store
func (m *mappingStore) Link(src, dst *Tree) {
	m.srcs[src] = dst
	m.dsts[dst] = src
}

// Has checks if a mapping exists in the store
func (m *mappingStore) Has(src, dst *Tree) bool {
	t, ok := m.srcs[src]
	if !ok {
		return false
	}

	return t == dst
}

// GetDst returns destination tree for the source
func (m *mappingStore) GetDst(src *Tree) (*Tree, bool) {
	t, ok := m.srcs[src]
	return t, ok
}

// GetSrc returns source tree for the destination
func (m *mappingStore) GetSrc(dst *Tree) (*Tree, bool) {
	t, ok := m.dsts[dst]
	return t, ok
}

// Size returns number of pair in the store
func (m *mappingStore) Size() int {
	return len(m.srcs)
}

func (m *mappingStore) ToList() []Mapping {
	list := make([]Mapping, len(m.srcs))
	i := 0
	for left, right := range m.srcs {
		list[i] = Mapping{left, right}
		i++
	}

	return list
}

type multiMapping struct {
	srcs map[*Tree]map[*Tree]bool
	dsts map[*Tree]map[*Tree]bool
}

func newMultiMapping() *multiMapping {
	return &multiMapping{
		srcs: make(map[*Tree]map[*Tree]bool),
		dsts: make(map[*Tree]map[*Tree]bool),
	}
}

func (m *multiMapping) Link(src, dst *Tree) {
	if _, ok := m.srcs[src]; !ok {
		m.srcs[src] = make(map[*Tree]bool)
	}
	m.srcs[src][dst] = true

	if _, ok := m.dsts[dst]; !ok {
		m.dsts[dst] = make(map[*Tree]bool)
	}
	m.dsts[dst][src] = true
}

func (m *multiMapping) IsSrcUnique(src *Tree) bool {
	s := m.srcs[src]
	if len(s) != 1 {
		return false
	}

	var d *Tree
	for k := range m.srcs[src] {
		d = k
	}

	return len(m.dsts[d]) == 1
}
