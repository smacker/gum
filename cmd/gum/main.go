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
	"text/template"

	"github.com/smacker/gum"
	"github.com/smacker/gum/golang"
	"github.com/smacker/gum/uast"
	bblfshUAST "gopkg.in/bblfsh/sdk.v2/uast"

	flags "github.com/jessevdk/go-flags"
	bblfsh "gopkg.in/bblfsh/client-go.v2"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
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
	for _, t := range gum.PreOrder(root) {
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

type webCommand struct {
	parseOptions
}

func (c *webCommand) Execute(args []string) error {
	if c.Parser != "bblfsh" {
		return fmt.Errorf("only bblfsh driver supports webdiff for now")
	}

	src, dst, err := c.parse()
	if err != nil {
		return err
	}

	mappings := gum.Match(src, dst)
	actions := gum.Patch(src, dst, mappings)
	srcGroups, dstGroups := c.treeGroups(actions, mappings)

	srcb, err := ioutil.ReadFile(c.Args.Src)
	if err != nil {
		return err
	}
	dstb, err := ioutil.ReadFile(c.Args.Dst)
	if err != nil {
		return err
	}

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}

	htmlf, err := ioutil.TempFile("", "gum_diff_*.html")
	if err != nil {
		return err
	}

	if err := t.Execute(htmlf, struct {
		SrcHTML string
		DstHTML string
	}{
		SrcHTML: c.genHTML(srcb, c.srcTags(src, srcGroups)),
		DstHTML: c.genHTML(dstb, c.dstTags(dst, dstGroups)),
	}); err != nil {
		return err
	}

	if err := htmlf.Close(); err != nil {
		return err
	}

	fmt.Println("html file:", htmlf.Name())
	if runtime.GOOS == "darwin" {
		_ = exec.Command("open", htmlf.Name()).Run()
	} else if runtime.GOOS == "linux" {
		_ = exec.Command("xdg-open", htmlf.Name()).Run()
	}

	return nil
}

func (c *webCommand) treeGroups(actions []*gum.Action, mappings []gum.Mapping) (map[string][]*gum.Tree, map[string][]*gum.Tree) {
	srcGroups := map[string][]*gum.Tree{
		"mv":  []*gum.Tree{},
		"del": []*gum.Tree{},
		"upd": []*gum.Tree{},
	}
	dstGroups := map[string][]*gum.Tree{
		"add": []*gum.Tree{},
		"mv":  []*gum.Tree{},
		"del": []*gum.Tree{},
		"upd": []*gum.Tree{},
	}

	for _, a := range actions {
		switch a.Type {
		case gum.Insert:
			dstGroups["add"] = append(dstGroups["add"], a.Node)
		case gum.InsertTree:
			dstGroups["add"] = append(dstGroups["add"], a.Node)
			for _, n := range gum.PreOrder(a.Node) {
				dstGroups["add"] = append(dstGroups["add"], n)
			}
		case gum.DeleteTree:
			srcGroups["del"] = append(srcGroups["del"], a.Node)
			for _, n := range gum.PreOrder(a.Node) {
				srcGroups["del"] = append(srcGroups["del"], n)
			}
		case gum.Delete:
			srcGroups["del"] = append(srcGroups["del"], a.Node)
		case gum.Update:
			srcGroups["upd"] = append(srcGroups["upd"], a.Node)
			dstGroups["upd"] = append(dstGroups["upd"], getDst(mappings, a.Node))
		case gum.Move:
			srcGroups["mv"] = append(srcGroups["mv"], a.Node)
			dstGroups["mv"] = append(dstGroups["mv"], getDst(mappings, a.Node))
		}
	}

	return srcGroups, dstGroups
}

func getDst(mappings []gum.Mapping, t *gum.Tree) *gum.Tree {
	for _, m := range mappings {
		if m[0] == t {
			return m[1]
		}
	}

	return nil
}

func (c *webCommand) srcTags(src *gum.Tree, treeGroups map[string][]*gum.Tree) *tags {
	tags := newTags()
	for _, t := range gum.PreOrder(src) {
		n := t.Meta.(nodes.Node)
		pos := bblfshUAST.PositionsOf(n)
		start := int(pos["start"].Offset)
		end := int(pos["end"].Offset)

		switch true {
		case inGroup(treeGroups["mv"], t):
			tags.add(start, end, "mv")
		case inGroup(treeGroups["upd"], t):
			tags.add(start, end, "upd")
		case inGroup(treeGroups["del"], t):
			tags.add(start, end, "del")
		}
	}

	return tags
}

func (c *webCommand) dstTags(dst *gum.Tree, treeGroups map[string][]*gum.Tree) *tags {
	tags := newTags()
	for _, t := range gum.PreOrder(dst) {
		n := t.Meta.(nodes.Node)
		pos := bblfshUAST.PositionsOf(n)
		start := int(pos["start"].Offset)
		end := int(pos["end"].Offset)

		switch true {
		case inGroup(treeGroups["mv"], t):
			tags.add(start, end, "mv")
		case inGroup(treeGroups["upd"], t):
			tags.add(start, end, "upd")
		case inGroup(treeGroups["add"], t):
			tags.add(start, end, "add")
		}
	}

	return tags
}

func (c *webCommand) genHTML(text []byte, tags *tags) string {
	var htmlb []byte
	var i int
	for _, ch := range text {
		if v, ok := tags.starts[i]; ok {
			for _, action := range v {
				htmlb = append(htmlb, []byte("<span class='"+action+"'>")...)
			}
		}
		if v, ok := tags.ends[i]; ok {
			for j := 0; j < v; j++ {
				htmlb = append(htmlb, []byte("</span>")...)
			}
		}
		htmlb = append(htmlb, ch)
		i++
	}
	if v, ok := tags.ends[i]; ok {
		for j := 0; j < v; j++ {
			htmlb = append(htmlb, []byte("</span>")...)
		}
	}

	return string(htmlb)
}

type tags struct {
	starts map[int][]string
	ends   map[int]int
}

func newTags() *tags {
	return &tags{
		starts: make(map[int][]string),
		ends:   make(map[int]int),
	}
}

func (ts *tags) add(start, end int, v string) {
	ts.starts[start] = append(ts.starts[start], v)
	ts.ends[end]++
}

func inGroup(trees []*gum.Tree, t *gum.Tree) bool {
	for _, i := range trees {
		if i == t {
			return true
		}
	}
	return false
}

const tpl = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>Webdiff</title>
    <style>
      .del {
        background: #ffeef0;
      }
      .add {
        background: #e6ffed;
	  }
	  .upd {
        background: #ffffd8;
	  }
      .mv {
        background: #efefef;
      }
    </style>
  </head>
  <body>
    <div style="display: flex;">
      <div style="width:50%;">
        <pre>{{ .SrcHTML }}</pre>
      </div>
      <div>
        <pre>{{ .DstHTML }}</pre>
      </div>
    </div>
  </body>
</html>
`

func main() {
	parser := flags.NewNamedParser("gum", flags.Default)

	parser.AddCommand("match", "parse and display matched nodes", "", &matchCommand{})
	parser.AddCommand("diff", "parse and display actions", "", &diffCommand{})
	parser.AddCommand("webdiff", "parse and show web diff", "", &webCommand{})

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
