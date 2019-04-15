package uast

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/smacker/gum"
	"github.com/stretchr/testify/assert"
	uastyml "gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

func TestToTree(t *testing.T) {
	assert := assert.New(t)

	b, err := ioutil.ReadFile("testdata/src.uast")
	assert.NoError(err)
	node, err := uastyml.Unmarshal(b)
	assert.NoError(err)

	src := ToTree(node)
	assert.Equal("CompilationUnit", src.Type)
	withLabel := src.Children[0].Children[1]
	assert.Equal("Modifier", withLabel.Type)
	assert.Equal("public", withLabel.Value)

	b, err = ioutil.ReadFile("testdata/dst.uast")
	assert.NoError(err)
	node, err = uastyml.Unmarshal(b)
	assert.NoError(err)
	dst := ToTree(node)

	treePrint(dst, 0)

	mappings := gum.Match(src, dst)
	// the number is different from gum_test due to different ast produced by bblfsh
	// 1. NumberLiteral doesn't have label (must be bug in bblfsh.TokenOf)
	// 2. SimpleType doesn't have label (which makes sense)
	assert.Len(mappings, 12)
	for _, m := range mappings {
		fmt.Println(m)
	}

	actions := gum.Patch(src, dst, mappings)
	// current code generates some trash, it's wrong number
	assert.Len(actions, 38)
	for _, a := range actions {
		fmt.Println(a)
	}
}

func treePrint(t *gum.Tree, tab int) {
	fmt.Println(strings.Repeat("-", tab), t)
	for _, c := range t.Children {
		treePrint(c, tab+1)
	}
}
