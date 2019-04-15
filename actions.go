package gum

func newInsert(node, parent *Tree, pos int) *Action {
	return &Action{Type: Insert, Node: node, Parent: parent, Pos: pos}
}

func newTreeInsert(node, parent *Tree, pos int) *Action {
	return &Action{Type: InsertTree, Node: node, Parent: parent, Pos: pos}
}

func newUpdate(node *Tree, value string) *Action {
	return &Action{Type: Update, Node: node, Value: value}
}

func newMove(node, parent *Tree, pos int) *Action {
	return &Action{Type: Move, Node: node, Parent: parent, Pos: pos}
}

func newDelete(node *Tree) *Action {
	return &Action{Type: Delete, Node: node}
}

func newTreeDelete(node *Tree) *Action {
	return &Action{Type: DeleteTree, Node: node}
}

type actionGenerator struct {
	origSrc *Tree
	newSrc  *Tree
	origDst *Tree

	origSrcTrees map[int]*Tree
	cpySrcTrees  map[int]*Tree

	origMappings *mappingStore
	// original mapping + link to fake nodes corresponding to nodes in dst tree that are missed in src tree
	newMappings *mappingStore

	dstInOrder map[*Tree]bool
	srcInOrder map[*Tree]bool

	lastID int
}

func newActionGenerator(src, dst *Tree, mappings []Mapping) *actionGenerator {
	g := &actionGenerator{}
	g.init(src, dst, mappings)

	return g
}

func (g *actionGenerator) init(src, dst *Tree, mappings []Mapping) {
	g.origSrc = src
	g.newSrc = src.clone()
	g.origDst = dst

	g.origSrcTrees = make(map[int]*Tree)
	for _, t := range getTrees(g.origSrc) {
		g.origSrcTrees[t.id] = t
	}
	g.cpySrcTrees = make(map[int]*Tree)
	for _, t := range getTrees(g.newSrc) {
		g.cpySrcTrees[t.id] = t
	}

	g.origMappings = newMappingStore()
	for _, m := range mappings {
		g.origMappings.Link(g.cpySrcTrees[m[0].id], m[1])
	}

	g.newMappings = newMappingStore()
	for _, m := range mappings {
		g.newMappings.Link(g.cpySrcTrees[m[0].id], m[1])
	}
}

func (g *actionGenerator) Generate() []*Action {
	srcFakeRoot := newFakeTree(g.newSrc)
	dstFakeRoot := newFakeTree(g.origDst)
	g.newSrc.parent = srcFakeRoot
	g.origDst.parent = dstFakeRoot

	actions := make([]*Action, 0)
	g.dstInOrder = make(map[*Tree]bool)
	g.srcInOrder = make(map[*Tree]bool)

	g.lastID = g.newSrc.size + 1
	g.newMappings.Link(srcFakeRoot, dstFakeRoot)

	for _, x := range breadthFirst(g.origDst) {
		// x - node of the original dst tree
		// w - corresponding node to x in the src tree
		// y - parent of the original dst tree
		// z - corresponding node to y in the src tree
		var w *Tree
		y := x.parent
		z, _ := g.newMappings.GetSrc(y)

		var hasDst bool
		w, hasDst = g.newMappings.GetSrc(x)
		if !hasDst {
			// insert new node if there is no such node in dst tree side of new mapping
			k := g.findPos(x)
			w = &Tree{id: g.newID(), parent: z}
			// use real node in the action
			ins := newInsert(x, g.origSrcTrees[z.id], k)
			actions = append(actions, ins)

			// id of fake node will return the real node
			g.origSrcTrees[w.id] = x
			// add fake node into new original tree
			g.newMappings.Link(w, x)
			// update parent with of the node in original tree with newly created node
			// it's safe because we keep clones in the src side of the new mapping
			z.addChild(k, w)
		} else if x != g.origDst { // x == origDst is a special case for the root of the tree
			if w.Value != x.Value {
				actions = append(actions, newUpdate(g.origSrcTrees[w.id], x.Value))
				// update the clone
				w.Value = x.Value
			}
			v := w.parent
			if z != v {
				k := g.findPos(x)
				mv := newMove(g.origSrcTrees[w.id], g.origSrcTrees[z.id], k)
				actions = append(actions, mv)
				// update the clone
				oldk := positionInParent(w)
				z.addChild(k, w)
				w.parent.removeChild(w.parent.Children[oldk])
				w.parent = z
			}
		}

		g.srcInOrder[w] = true
		g.dstInOrder[x] = true
		actions = append(actions, g.alignChildren(w, x)...)
	}

	for _, w := range postOrder(g.newSrc) {
		if _, ok := g.newMappings.GetDst(w); !ok {
			actions = append(actions, newDelete(g.origSrcTrees[w.id]))
		}
	}

	return g.simplify(actions)
}

func (g *actionGenerator) alignChildren(w, x *Tree) []*Action {
	var actions []*Action

	for _, c := range w.Children {
		delete(g.srcInOrder, c)
	}
	for _, c := range x.Children {
		delete(g.dstInOrder, c)
	}

	// list of children of src node that has mapping in dst tree
	s1 := make([]*Tree, 0)
	for _, c := range w.Children {
		if d, ok := g.newMappings.GetDst(c); ok {
			if getChildPosition(x, d) != -1 {
				s1 = append(s1, c)
			}
		}
	}

	// list of children of dst node that has mapping in src tree
	s2 := make([]*Tree, 0)
	for _, c := range x.Children {
		if s, ok := g.newMappings.GetSrc(c); ok {
			if getChildPosition(w, s) != -1 {
				s2 = append(s2, c)
			}
		}
	}

	lcs := g.makeLcs(s1, s2)
	for _, m := range lcs {
		g.srcInOrder[m[0]] = true
		g.dstInOrder[m[1]] = true
	}

	for _, a := range s1 {
		for _, b := range s2 {
			if !g.origMappings.Has(a, b) {
				continue
			}
			hasMapping := false
			for _, m := range lcs {
				if m[0] == a && m[1] == b {
					hasMapping = true
					break
				}
			}
			if !hasMapping {
				k := g.findPos(b)
				mv := newMove(g.origSrcTrees[a.id], g.origSrcTrees[w.id], k)
				actions = append(actions, mv)

				oldk := positionInParent(a)
				w.addChild(k, a)
				if k < oldk {
					oldk++
				}
				a.parent.removeChild(w.parent.Children[oldk])
				a.parent = w
				g.srcInOrder[a] = true
				g.dstInOrder[b] = true
			}

		}
	}

	return actions
}

func (g *actionGenerator) makeLcs(x, y []*Tree) []Mapping {
	lcs := make([]Mapping, 0)

	m := len(x)
	n := len(y)

	opt := make([][]int, m+1)
	for i := range opt {
		opt[i] = make([]int, n+1)
	}

	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if g.newMappings.dsts[y[j]] == x[i] {
				opt[i][j] = opt[i+1][j+1] + 1
			} else {
				max := opt[i+1][j]
				if opt[i][j+1] > max {
					max = opt[i][j+1]
				}
				opt[i][j] = max
			}
		}
	}

	i := 0
	j := 0
	for i < m && j < n {
		if g.newMappings.dsts[y[j]] == x[i] {
			lcs = append(lcs, Mapping{x[i], y[j]})
			i++
			j++
		} else if opt[i+1][j] >= opt[i][j+1] {
			i++
		} else {
			j++
		}
	}

	return lcs
}

func (g *actionGenerator) findPos(x *Tree) int {
	y := x.parent
	siblings := y.Children

	for _, c := range siblings {
		if _, ok := g.dstInOrder[c]; ok {
			if c == x {
				return 0
			}
			break
		}
	}

	xpos := positionInParent(x)
	var v *Tree
	for i := 0; i < xpos; i++ {
		c := siblings[i]
		if _, ok := g.dstInOrder[c]; ok {
			v = c
		}
	}

	if v == nil {
		return 0
	}

	u, _ := g.newMappings.GetSrc(v)
	upos := positionInParent(u)

	return upos + 1
}

func containsAll(m map[*Tree]*Action, nodes ...*Tree) bool {
	for _, n := range nodes {
		_, ok := m[n]
		if !ok {
			return false
		}
	}

	return true
}

func removeAction(actions []*Action, act *Action) []*Action {
	newActions := make([]*Action, 0)
	for _, a := range actions {
		if a == act {
			continue
		}
		newActions = append(newActions, a)
	}

	return newActions
}

func replaceAction(actions []*Action, old, new *Action) []*Action {
	for i, a := range actions {
		if a == old {
			return append(actions[:i], append([]*Action{new}, actions[i+1:]...)...)
		}
	}

	return actions
}

func (g *actionGenerator) simplify(actions []*Action) []*Action {
	addedTrees := make(map[*Tree]*Action)
	deletedTrees := make(map[*Tree]*Action)

	for _, a := range actions {
		switch a.Type {
		case Insert:
			addedTrees[a.Node] = a
		case Delete:
			deletedTrees[a.Node] = a
		}
	}

	for t, a := range addedTrees {
		_, ok := addedTrees[t.parent]
		if ok && containsAll(addedTrees, getDescendants(t)...) {
			actions = removeAction(actions, a)
		} else {
			if len(t.Children) > 0 && containsAll(addedTrees, getDescendants(t)...) {
				ti := newTreeInsert(a.Node, a.Parent, a.Pos)
				actions = replaceAction(actions, a, ti)
			}
		}
	}

	for t, a := range deletedTrees {
		_, ok := deletedTrees[t.parent]
		if ok && containsAll(deletedTrees, getDescendants(t)...) {
			actions = removeAction(actions, a)
		} else {
			if len(t.Children) > 0 && containsAll(deletedTrees, getDescendants(t)...) {
				td := newTreeDelete(a.Node)
				actions = replaceAction(actions, a, td)
			}
		}
	}

	return actions
}

func (g *actionGenerator) newID() int {
	id := g.lastID
	g.lastID++
	return id
}

func newFakeTree(t *Tree) *Tree {
	return &Tree{Children: []*Tree{t}}
}
