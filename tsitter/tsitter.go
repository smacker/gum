package tsitter

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/gum"
)

// let's just hard-code types for golang for now
// number + all identifiers
var goTokenTypes = map[string]bool{
	"identifier":                 true,
	"blank_identifier":           true,
	"raw_string_literal":         true,
	"int_literal":                true,
	"float_literal":              true,
	"imaginary_literal":          true,
	"rune_literal":               true,
	"comment":                    true,
	"composite_literal":          true,
	"package_identifier":         true,
	"field_identifier":           true,
	"type_identifier":            true,
	"label_name":                 true,
	"interpreted_string_literal": true,
	"func_literal":               true,
}

// Get all symbols for a language
//
// gl := golang.GetLanguage()
// for i := uint32(0); i < gl.SymbolCount(); i++ {
// 	if gl.SymbolType(sitter.Symbol(i)) != sitter.SymbolTypeRegular {
// 		continue
// 	}

// 	fmt.Println(gl.SymbolName(sitter.Symbol(i)))
// }

// ToTree converts bblfsh.Node to gum.Tree
func ToTree(n *sitter.Node, source []byte) *gum.Tree {
	t := toTree(n, source)
	t.Refresh()
	return t
}

func toTree(n *sitter.Node, source []byte) *gum.Tree {
	var children []*gum.Tree
	for i := uint32(0); i < n.NamedChildCount(); i++ {
		children = append(children, toTree(n.NamedChild(int(i)), source))
	}
	var value string
	if _, ok := goTokenTypes[n.Type()]; ok {
		value = string(source[n.StartByte():n.EndByte()])
	}
	tree := &gum.Tree{
		Type:     n.Type(),
		Meta:     n,
		Value:    value,
		Children: children,
	}

	return tree
}
