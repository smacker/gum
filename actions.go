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

// Generates edit script
//
// Algorithm is based on the paper “Change Detection in Hierarchically Structured Information”
// by S.S. Chawathe, A. Rajaraman, H. Garcia-Molina, and J. Widom.
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

	skipSimplify bool
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
			// Insert phase
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
			// update parent of the node in original tree with newly created node
			// it's safe because we keep clones in the src side of the new mapping
			z.addChild(k, w)
		} else if x != g.origDst { // x == origDst is a special case for the root of the tree
			// Update phase
			if w.Value != x.Value {
				actions = append(actions, newUpdate(g.origSrcTrees[w.id], x.Value))
				// update the clone
				w.Value = x.Value
			}
			// Move phase
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

		// FIXME: looks like srcInOrder is never used
		g.srcInOrder[w] = true
		g.dstInOrder[x] = true

		actions = append(actions, g.alignChildren(w, x)...)
	}

	// Delete phase
	for _, w := range postOrder(g.newSrc) {
		if _, ok := g.newMappings.GetDst(w); !ok {
			actions = append(actions, newDelete(g.origSrcTrees[w.id]))
		}
	}

	if g.skipSimplify {
		return actions
	}

	return g.simplify(actions)
}

// children of w and x are misaligned
// when mapped children of w have different order in x
//
// In short: Compute LCS for mapped children, fix them,
// move the rest of mapped children relatively to already fixed
func (g *actionGenerator) alignChildren(w, x *Tree) []*Action {
	var actions []*Action

	// mark all children of w and x as out of order
	for _, c := range w.Children {
		delete(g.srcInOrder, c)
	}
	for _, c := range x.Children {
		delete(g.dstInOrder, c)
	}

	// children of src node that are mappend and belong to a partner node in dst
	s1 := make([]*Tree, 0)
	for _, c := range w.Children {
		if d, ok := g.newMappings.GetDst(c); ok {
			if getChildPosition(x, d) != -1 {
				s1 = append(s1, c)
			}
		}
	}

	// children of dst node that are mappend and belong to a partner node in src
	s2 := make([]*Tree, 0)
	for _, c := range x.Children {
		if s, ok := g.newMappings.GetSrc(c); ok {
			if getChildPosition(w, s) != -1 {
				s2 = append(s2, c)
			}
		}
	}

	// to align children there is more than one sequence of moves
	// we want the minimal number of moves
	// use longest common subsequence algorithm for that
	lcs := g.makeLcs(s1, s2)
	// put all children from the longest subsequence into "in order"
	for _, m := range lcs {
		g.srcInOrder[m[0]] = true
		g.dstInOrder[m[1]] = true
	}

	for _, a := range s1 {
		for _, b := range s2 {
			/// FIXME: looks like no need, we already checked during building s1 & s1
			// we checked in newMappings, not origMappings but newMappings contains all origMappings
			if !g.origMappings.Has(a, b) {
				continue
			}
			// skip children that are in LCS
			ordered := false
			for _, m := range lcs {
				if m[0] == a && m[1] == b {
					ordered = true
					break
				}
			}
			if ordered {
				continue
			}

			// make a move relatively to the siblings "in order"
			k := g.findPos(b)
			mv := newMove(g.origSrcTrees[a.id], g.origSrcTrees[w.id], k)
			actions = append(actions, mv)

			// apply move
			oldk := positionInParent(a)
			a.parent.removeChild(a.parent.Children[oldk])
			if k > oldk {
				k--
			}
			w.addChild(k, a)
			a.parent = w

			// Mark a and b "in order"
			g.srcInOrder[a] = true
			g.dstInOrder[b] = true
		}
	}

	return actions
}

// make longest common subsequence of lists of children
func (g *actionGenerator) makeLcs(x, y []*Tree) []Mapping {
	lcs := make([]Mapping, 0)

	m := len(x)
	n := len(y)

	// prepate LCS table
	opt := make([][]int, m+1)
	for i := range opt {
		opt[i] = make([]int, n+1)
	}

	// fill LCS table with the lengths
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

	// traceback to get the subsequence
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

// finds where to put a node from dst tree into src tree
// using relative positions of siblings that are known to be in order
func (g *actionGenerator) findPos(x *Tree) int {
	y := x.parent
	siblings := y.Children

	// If x is the leftmost child of y that is marked "in order" return 1
	for _, c := range siblings {
		if _, ok := g.dstInOrder[c]; ok {
			if c == x {
				return 0
			}
			break
		}
	}

	// Find v in T2 where v is the rightmost sibling of x that is to the left of x and is marked "in order"
	var v *Tree
	xpos := positionInParent(x)
	for i := xpos - 1; i >= 0; i-- {
		c := siblings[i]
		if _, ok := g.dstInOrder[c]; ok {
			v = c
			break
		}
	}

	// This case is not described in the paper (assume it is the first node)
	if v == nil {
		return 0
	}

	// Let u be the partner of v in T1.
	u, ok := g.newMappings.GetSrc(v)
	if !ok {
		return 0
	}
	// Suppose u is the ith child of its parent (counting from left to right) that is marked "in order"
	// return i+1
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
	newActions := make([]*Action, len(actions))
	for i, a := range actions {
		newActions[i] = a
	}

	lastType := Insert
	seqTrees := make(map[*Tree]*Action)
	for i, a := range actions {
		// TODO: update doesn't change the structure of a tree
		// it's safe to squash actions as long as update is the last operation

		if i > 0 && a.Type != lastType && len(seqTrees) > 0 {
			if lastType == Insert {
				newActions = simplifyInsert(seqTrees, newActions)
			}
			if lastType == Delete {
				newActions = simplifyDelete(seqTrees, newActions)
			}

			seqTrees = make(map[*Tree]*Action)
		}

		if a.Type == Insert || a.Type == Delete {
			seqTrees[a.Node] = a
		}

		lastType = a.Type
	}

	if len(seqTrees) > 0 {
		if lastType == Insert {
			newActions = simplifyInsert(seqTrees, newActions)
		}
		if lastType == Delete {
			newActions = simplifyDelete(seqTrees, newActions)
		}
	}

	return newActions
}

func simplifyInsert(trees map[*Tree]*Action, actions []*Action) []*Action {
	var actionToReplace *Action
	height := 1
	for t, a := range trees {
		// find the biggest tree with all descendants inserted
		if containsAll(trees, getDescendants(t)...) && height < t.height {
			actionToReplace = a
			height = t.height
		}
	}

	if actionToReplace != nil {
		a := actionToReplace
		ti := newTreeInsert(a.Node, a.Parent, a.Pos)
		actions = replaceAction(actions, a, ti)
		for _, t := range getDescendants(a.Node) {
			actions = removeAction(actions, trees[t])
		}
	}

	return actions
}

func simplifyDelete(trees map[*Tree]*Action, actions []*Action) []*Action {
	var actionToReplace *Action
	height := 1
	for t, a := range trees {
		// find the biggest tree with all descendants inserted
		if containsAll(trees, getDescendants(t)...) && height < t.height {
			actionToReplace = a
			height = t.height
		}
	}

	if actionToReplace != nil {
		a := actionToReplace
		td := newTreeDelete(a.Node)
		actions = replaceAction(actions, a, td)
		for _, t := range getDescendants(a.Node) {
			actions = removeAction(actions, trees[t])
		}
	}

	return actions
}

func (g *actionGenerator) newID() int {
	id := g.lastID
	g.lastID++
	return id + 1
}

func newFakeTree(t *Tree) *Tree {
	return &Tree{Children: []*Tree{t}}
}
