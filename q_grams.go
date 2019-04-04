package gum

import (
	"math"
	"strings"
)

// FIXME: handle unicode correctly or replace it with 3rd party lib

type tokenizer interface {
	Tokenize(s string) []string
}

type qGram struct {
	q int
}

func newQGram(q int) *qGram {
	return &qGram{q}
}

func (q *qGram) Tokenize(s string) []string {
	if s == "" {
		return nil
	}
	if len(s) <= q.q {
		return []string{s}
	}

	res := make([]string, len(s)-q.q+1)
	lastStart := len(s) - q.q
	for i := 0; i <= lastStart; i++ {
		res[i] = s[i : i+q.q]
	}

	return res
}

type qGramExtended struct {
	t            *qGram
	startPadding string
	endPadding   string
}

func newQGramExtended(q int, startPadding, endPadding string) *qGramExtended {
	return &qGramExtended{newQGram(q), strings.Repeat(startPadding, q-1), strings.Repeat(endPadding, q-1)}
}

func (q *qGramExtended) Tokenize(s string) []string {
	if s == "" {
		return nil
	}

	return q.t.Tokenize(q.startPadding + s + q.endPadding)
}

type blockDistance struct {
	t tokenizer
}

func qGramsDistance() *blockDistance {
	return &blockDistance{newQGramExtended(3, "#", "#")}
}

func (d *blockDistance) Compare(a, b string) float64 {
	return d.compare(d.t.Tokenize(a), d.t.Tokenize(b))
}

func (d *blockDistance) compare(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	return 1 - float64(d.distance(a, b))/float64(len(a)+len(b))
}

func (d *blockDistance) distance(a, b []string) int {
	distance := 0

	unionSet := make(map[string]bool)
	for _, token := range a {
		unionSet[token] = true
	}
	for _, token := range b {
		unionSet[token] = true
	}

	for token := range unionSet {
		frequencyInA := d.count(token, a)
		frequencyInB := d.count(token, b)

		distance += int(math.Abs(float64(frequencyInA - frequencyInB)))
	}

	return distance
}

func (d *blockDistance) count(token string, list []string) int {
	i := 0
	for _, t := range list {
		if t == token {
			i++
		}
	}

	return i
}
