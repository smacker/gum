package golang

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/smacker/gum"
)

// ToTree converts ast.File to gum.Tree
func ToTree(f *ast.File) *gum.Tree {
	t := toTree(f)
	t.Refresh()
	return t
}

func toTree(node ast.Node) *gum.Tree {
	var token string
	var children []*gum.Tree

	switch n := node.(type) {
	case *ast.Comment:
		token = n.Text
	case *ast.CommentGroup:
		for _, c := range n.List {
			children = append(children, toTree(c))
		}
	case *ast.Field:
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		for _, x := range n.Names {
			children = append(children, toTree(x))
		}
		children = append(children, toTree(n.Type))
		if n.Tag != nil {
			children = append(children, toTree(n.Tag))
		}
		if n.Comment != nil {
			children = append(children, toTree(n.Comment))
		}
	case *ast.FieldList:
		for _, f := range n.List {
			children = append(children, toTree(f))
		}
	// Expressions
	case *ast.BadExpr:
		// nothing to do
	case *ast.Ident:
		token = n.Name
	case *ast.BasicLit:
		token = n.Value
	case *ast.Ellipsis:
		if n.Elt != nil {
			children = append(children, toTree(n.Elt))
		}
	case *ast.FuncLit:
		children = append(children, toTree(n.Type))
		children = append(children, toTree(n.Body))
	case *ast.CompositeLit:
		if n.Type != nil {
			children = append(children, toTree(n.Type))
		}
		for _, x := range n.Elts {
			children = append(children, toTree(x))
		}
	case *ast.ParenExpr:
		children = append(children, toTree(n.X))
	case *ast.SelectorExpr:
		children = append(children, toTree(n.X))
		children = append(children, toTree(n.Sel))
	case *ast.IndexExpr:
		children = append(children, toTree(n.X))
		children = append(children, toTree(n.Index))
	case *ast.SliceExpr:
		children = append(children, toTree(n.X))
		if n.Low != nil {
			children = append(children, toTree(n.Low))
		}
		if n.High != nil {
			children = append(children, toTree(n.High))
		}
		if n.Max != nil {
			children = append(children, toTree(n.Max))
		}
	case *ast.TypeAssertExpr:
		children = append(children, toTree(n.X))
		if n.Type != nil {
			children = append(children, toTree(n.Type))
		}
	case *ast.CallExpr:
		children = append(children, toTree(n.Fun))
		for _, x := range n.Args {
			children = append(children, toTree(x))
		}
	case *ast.StarExpr:
		children = append(children, toTree(n.X))
	case *ast.UnaryExpr:
		token = n.Op.String()
		children = append(children, toTree(n.X))
	case *ast.BinaryExpr:
		token = n.Op.String()
		children = append(children, toTree(n.X))
		children = append(children, toTree(n.Y))
	case *ast.KeyValueExpr:
		children = append(children, toTree(n.Key))
		children = append(children, toTree(n.Value))
	// Types
	case *ast.ArrayType:
		if n.Len != nil {
			children = append(children, toTree(n.Len))
		}
		children = append(children, toTree(n.Elt))
	case *ast.StructType:
		children = append(children, toTree(n.Fields))
	case *ast.FuncType:
		if n.Params != nil {
			children = append(children, toTree(n.Params))
		}
		if n.Results != nil {
			children = append(children, toTree(n.Results))
		}
	case *ast.InterfaceType:
		children = append(children, toTree(n.Methods))
	case *ast.MapType:
		children = append(children, toTree(n.Key))
		children = append(children, toTree(n.Value))
	case *ast.ChanType:
		children = append(children, toTree(n.Value))
	// Statements
	case *ast.BadStmt:
		// nothing to do
	case *ast.DeclStmt:
		children = append(children, toTree(n.Decl))
	case *ast.EmptyStmt:
		// nothing to do
	case *ast.LabeledStmt:
		children = append(children, toTree(n.Label))
		children = append(children, toTree(n.Stmt))
	case *ast.ExprStmt:
		children = append(children, toTree(n.X))
	case *ast.SendStmt:
		children = append(children, toTree(n.Chan))
		children = append(children, toTree(n.Value))
	case *ast.IncDecStmt:
		token = n.Tok.String()
		children = append(children, toTree(n.X))
	case *ast.AssignStmt:
		token = n.Tok.String()
		for _, x := range n.Lhs {
			children = append(children, toTree(x))
		}
		for _, x := range n.Rhs {
			children = append(children, toTree(x))
		}
	case *ast.GoStmt:
		children = append(children, toTree(n.Call))
	case *ast.DeferStmt:
		children = append(children, toTree(n.Call))
	case *ast.ReturnStmt:
		for _, x := range n.Results {
			children = append(children, toTree(x))
		}
	case *ast.BranchStmt:
		token = n.Tok.String()
		if n.Label != nil {
			children = append(children, toTree(n.Label))
		}
	case *ast.BlockStmt:
		for _, x := range n.List {
			children = append(children, toTree(x))
		}
	case *ast.IfStmt:
		if n.Init != nil {
			children = append(children, toTree(n.Init))
		}
		children = append(children, toTree(n.Cond))
		children = append(children, toTree(n.Body))
		if n.Else != nil {
			children = append(children, toTree(n.Else))
		}
	case *ast.CaseClause:
		for _, x := range n.List {
			children = append(children, toTree(x))
		}
		for _, x := range n.Body {
			children = append(children, toTree(x))
		}
	case *ast.SwitchStmt:
		if n.Init != nil {
			children = append(children, toTree(n.Init))
		}
		if n.Tag != nil {
			children = append(children, toTree(n.Tag))
		}
		children = append(children, toTree(n.Body))
	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			children = append(children, toTree(n.Init))
		}
		children = append(children, toTree(n.Assign))
		children = append(children, toTree(n.Body))
	case *ast.CommClause:
		if n.Comm != nil {
			children = append(children, toTree(n.Comm))
		}
		for _, x := range n.Body {
			children = append(children, toTree(x))
		}
	case *ast.SelectStmt:
		children = append(children, toTree(n.Body))
	case *ast.ForStmt:
		if n.Init != nil {
			children = append(children, toTree(n.Init))
		}
		if n.Cond != nil {
			children = append(children, toTree(n.Cond))
		}
		if n.Post != nil {
			children = append(children, toTree(n.Post))
		}
		children = append(children, toTree(n.Body))
	case *ast.RangeStmt:
		token = n.Tok.String()
		if n.Key != nil {
			children = append(children, toTree(n.Key))
		}
		if n.Value != nil {
			children = append(children, toTree(n.Value))
		}
		children = append(children, toTree(n.X))
		children = append(children, toTree(n.Body))
	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		if n.Name != nil {
			children = append(children, toTree(n.Name))
		}
		children = append(children, toTree(n.Path))
		if n.Comment != nil {
			children = append(children, toTree(n.Comment))
		}
	case *ast.ValueSpec:
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		for _, x := range n.Names {
			children = append(children, toTree(x))
		}
		if n.Type != nil {
			children = append(children, toTree(n.Type))
		}
		for _, x := range n.Values {
			children = append(children, toTree(x))
		}
		if n.Comment != nil {
			children = append(children, toTree(n.Comment))
		}
	case *ast.TypeSpec:
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		children = append(children, toTree(n.Name))
		children = append(children, toTree(n.Type))
		if n.Comment != nil {
			children = append(children, toTree(n.Comment))
		}
	case *ast.BadDecl:
		// nothing to do
	case *ast.GenDecl:
		token = n.Tok.String()
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		for _, s := range n.Specs {
			children = append(children, toTree(s))
		}
	case *ast.FuncDecl:
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		if n.Recv != nil {
			children = append(children, toTree(n.Recv))
		}
		children = append(children, toTree(n.Name))
		children = append(children, toTree(n.Type))
		if n.Body != nil {
			children = append(children, toTree(n.Body))
		}
	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			children = append(children, toTree(n.Doc))
		}
		children = append(children, toTree(n.Name))
		for _, x := range n.Decls {
			children = append(children, toTree(x))
		}
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes
	case *ast.Package:
		for _, f := range n.Files {
			children = append(children, toTree(f))
		}
	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	tree := &gum.Tree{
		Type:     reflect.TypeOf(node).Elem().Name(),
		Value:    token,
		Children: children,
	}

	return tree
}
