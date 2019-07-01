package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/smacker/gum"
	"github.com/smacker/gum/golang"
	"github.com/smacker/gum/uast"

	flags "github.com/jessevdk/go-flags"
	bblfsh "gopkg.in/bblfsh/client-go.v2"
)

type parseOptions struct {
	Parser string `short:"p" long:"parser" default:"bblfsh" choice:"bblfsh" choice:"go"`
	Args   struct {
		Src string
		Dst string
	} `positional-args:"yes" required:"yes"`
}

func (p *parseOptions) parse() (*gum.Tree, *gum.Tree, error) {
	src, err := parseFile(p.Args.Src, p.Parser)
	if err != nil {
		return nil, nil, err
	}
	dst, err := parseFile(p.Args.Dst, p.Parser)
	if err != nil {
		return nil, nil, err
	}

	return src, dst, nil
}

func parseFile(path string, parserName string) (*gum.Tree, error) {
	switch parserName {
	case "go":
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", string(b), parser.ParseComments)
		return golang.ToTree(f), nil
	case "bblfsh":
		client, err := bblfsh.NewClient("0.0.0.0:9432")
		if err != nil {
			return nil, fmt.Errorf("can't connect to bblfsh: %s", err)
		}
		res, _, err := client.
			NewParseRequest().
			Mode(bblfsh.Annotated).
			ReadFile(path).
			UAST()
		if err != nil {
			return nil, fmt.Errorf("can't parse the file %s: %s", path, err)
		}
		return uast.ToTree(res), nil
	default:
		return nil, fmt.Errorf("unknown parser %s", parserName)
	}
}

type matchCommand struct {
	parseOptions
	Mode string `short:"m" long:"mode" choice:"text" choice:"dot" choice:"png"`
}

func (c *matchCommand) Execute(args []string) error {
	if c.Mode == "" {
		_, err := exec.LookPath("dot")
		if err == nil {
			c.Mode = "png"
		} else {
			c.Mode = "text"
		}
	}

	src, dst, err := c.parse()
	if err != nil {
		return err
	}

	mappings := gum.Match(src, dst)

	switch c.Mode {
	case "text":
		for _, m := range mappings {
			fmt.Println(m[0].String())
		}
		return nil
	case "dot":
		graph(os.Stdout, src, dst, mappings)
		return nil
	case "png":
		dotf, err := ioutil.TempFile("", "gum_*.dot")
		if err != nil {
			return err
		}
		defer os.Remove(dotf.Name())
		graph(dotf, src, dst, mappings)
		if err := dotf.Close(); err != nil {
			return err
		}
		pngf, err := os.Create(strings.TrimSuffix(dotf.Name(), ".dot") + ".png")
		if err != nil {
			return err
		}
		if err := pngf.Close(); err != nil {
			return err
		}
		if err := dotToPng(dotf.Name(), pngf.Name()); err != nil {
			return err
		}

		fmt.Println("png file:", pngf.Name())
		if runtime.GOOS == "darwin" {
			_ = exec.Command("open", pngf.Name()).Run()
		} else if runtime.GOOS == "linux" {
			_ = exec.Command("xdg-open", pngf.Name()).Run()
		}

		return nil
	default:
		return fmt.Errorf("unknown mode %s", c.Mode)
	}
}

func graph(w io.Writer, src, dst *gum.Tree, ms []gum.Mapping) {
	fmt.Fprintln(w, "digraph G {")
	fmt.Fprintln(w, "node [style=filled];")
	fmt.Fprintln(w, "subgraph cluster_src {")
	writeTree(w, src, ms, "src")
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "subgraph cluster_dst {")
	writeTree(w, dst, ms, "dst")
	fmt.Fprintln(w, "}")
	for _, m := range ms {
		fmt.Fprintf(w, "%s -> %s [style=dashed]\n;", getDotID(m[0], "src"), getDotID(m[1], "dst"))
	}
	fmt.Fprintln(w, "}")
}

func writeTree(w io.Writer, root *gum.Tree, m []gum.Mapping, prefix string) {
	for _, t := range preOrder(root) {
		fillColor := "red"
		if inMapping(m, t) {
			fillColor = "blue"
		}
		fmt.Fprintf(w, "%s [label=\"%s\", color=%s];\n", getDotID(t, prefix), getDotLabel(t), fillColor)
		if t.GetParent() != nil {
			fmt.Fprintf(w, "%s -> %s;\n", getDotID(t.GetParent(), prefix), getDotID(t, prefix))
		}
	}
}

func getDotID(t *gum.Tree, prefix string) string {
	return fmt.Sprintf(prefix+"_%d", t.GetID())
}

func getDotLabel(t *gum.Tree) string {
	label := toPrettyString(t)
	label = strings.Replace(label, `"`, "", -1)
	label = strings.Replace(label, `\s`, "", -1)
	if len(label) > 30 {
		label = label[:30]
	}

	return label
}

func toPrettyString(t *gum.Tree) string {
	if t.Value == "" {
		return t.Type
	}
	return fmt.Sprintf("%s: %s", t.Type, t.Value)
}

func inMapping(ms []gum.Mapping, t *gum.Tree) bool {
	for _, m := range ms {
		if m[0] == t || m[1] == t {
			return true
		}
	}
	return false
}

func dotToPng(dotfile string, outfile string) error {
	return exec.Command("dot", "-Tpng", dotfile, "-o", outfile).Run()
}

type diffCommand struct {
	parseOptions
}

func (c *diffCommand) Execute(args []string) error {
	src, dst, err := c.parse()
	if err != nil {
		return err
	}

	mappings := gum.Match(src, dst)
	actions := gum.Patch(src, dst, mappings)

	matchers := make([]*jsonMatch, len(mappings))
	for i, m := range mappings {
		matchers[i] = &jsonMatch{Src: m[0].GetID(), Dst: m[1].GetID()}
	}

	jsonActions := make([]*jsonAction, len(actions))
	for i, a := range actions {
		parent := 0
		if a.Parent != nil {
			parent = a.Parent.GetID()
		}
		jsonActions[i] = &jsonAction{
			Action: typeToStr[a.Type],
			Tree:   a.Node.GetID(),
			Parent: parent,
			At:     a.Pos,
			Label:  a.Value,
		}
	}

	b, err := json.MarshalIndent(struct {
		Matches []*jsonMatch  `json:"matches"`
		Actions []*jsonAction `json:"actions"`
	}{matchers, jsonActions}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}

var typeToStr = map[gum.Operation]string{
	gum.Delete:     "delete",
	gum.DeleteTree: "delete-tree",
	gum.Insert:     "insert",
	gum.InsertTree: "insert-tree",
	gum.Update:     "update",
	gum.Move:       "move",
}

type jsonMatch struct {
	Src int `json:"src"`
	Dst int `json:"dest"`
}

type jsonAction struct {
	Action string `json:"action"`
	Tree   int    `json:"tree"`
	Parent int    `json:"parent,omitempty"`
	At     int    `json:"at,omitempty"`
	Label  string `json:"label,omitempty"`
}

func main() {
	parser := flags.NewNamedParser("gum", flags.Default)

	parser.AddCommand("match", "parse and display matched nodes", "", &matchCommand{})
	parser.AddCommand("diff", "parse and display actions", "", &diffCommand{})

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func preOrder(t *gum.Tree) []*gum.Tree {
	var trees []*gum.Tree

	trees = append(trees, t)
	for _, c := range t.Children {
		trees = append(trees, preOrder(c)...)
	}

	return trees
}
