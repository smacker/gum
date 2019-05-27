package gum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Requires running testdata/process_samples.sh
func TestSamples(t *testing.T) {
	if _, err := os.Stat("testdata/parsed/samples"); os.IsNotExist(err) {
		t.Skip("directory with processed samples doesn't exist")
	}

	var samples [][3]string
	err := filepath.Walk("testdata/parsed/samples", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, "_diff.") {
			samples = append(samples, [3]string{
				strings.Replace(path, "_diff.", "_v0.", 1),
				strings.Replace(path, "_diff.", "_v1.", 1),
				path,
			})
		}
		return nil
	})
	require.NoError(t, err)

	for _, sample := range samples {
		fmt.Println("checking ", sample[2])
		src, dst := readFixtures(sample[0], sample[1])

		var idMappings [][2]int
		mapping := Match(src, dst)
		for _, m := range mapping {
			idMappings = append(idMappings, [2]int{m[0].GetID(), m[1].GetID()})
		}
		sort.Slice(idMappings, func(i, j int) bool { return idMappings[i][0] < idMappings[j][0] })

		diff, err := parseGumTreeDiff(sample[2])
		if err != nil {
			panic(err)
		}
		var diffIDMappings [][2]int
		for _, m := range diff.Matches {
			diffIDMappings = append(diffIDMappings, [2]int{m.Src, m.Dst})
		}
		sort.Slice(diffIDMappings, func(i, j int) bool { return diffIDMappings[i][0] < diffIDMappings[j][0] })

		require.Equal(t, idMappings, diffIDMappings)
	}
}

func parseGumTreeDiff(path string) (*diff, error) {
	diffJSON, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var d diff
	if err := json.Unmarshal(diffJSON, &d); err != nil {
		return nil, err
	}

	return &d, nil
}

type matchDiffItem struct {
	Src int `json:"src"`
	Dst int `json:"dest"`
}

type diff struct {
	Matches []matchDiffItem `json:"matches"`
}

func debugSampleFailure(mapping []Mapping) {
	makeLabel := func(t *Tree) string {
		if t.Value != "" {
			return t.Type + ": " + t.Value
		}
		return t.Type
	}

	for _, m := range mapping {
		fmt.Println("Match " +
			makeLabel(m[0]) + "(" + strconv.Itoa(m[0].GetID()) + ") to " +
			makeLabel(m[1]) + "(" + strconv.Itoa(m[1].GetID()) + ")")
	}
}
