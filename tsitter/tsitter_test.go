package tsitter

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/gum"
	"github.com/stretchr/testify/assert"
)

func TestToTree(t *testing.T) {
	assert := assert.New(t)

	b, err := ioutil.ReadFile("testdata/src.go")
	assert.NoError(err)
	node := sitter.Parse(b, golang.GetLanguage())
	assert.NoError(err)

	src := ToTree(node, b)
	assert.Equal("source_file", src.Type)
	withLabel := src.Children[0].Children[0]
	assert.Equal("package_identifier", withLabel.Type)
	assert.Equal("main", withLabel.Value)

	b, err = ioutil.ReadFile("testdata/dst.go")
	assert.NoError(err)
	node = sitter.Parse(b, golang.GetLanguage())
	assert.NoError(err)
	dst := ToTree(node, b)

	treePrint(dst, 0)

	// check that mapping works
	mappings := gum.Match(src, dst)
	assert.Len(mappings, 16)
	for _, m := range mappings {
		fmt.Println(m)
	}

	// check that patching works
	actions := gum.Patch(src, dst, mappings)
	assert.Len(actions, 1)
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
