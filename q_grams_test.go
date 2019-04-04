package gum

import (
	"fmt"
	"testing"
)

type tCase struct {
	Input    string
	Expected []string
}

func TestQGram1(t *testing.T) {
	cases := []tCase{
		{"", []string{}},
		{"1", []string{"1"}},
		{"12", []string{"1", "2"}},
		{"123456789", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}},
		{"123456789123456789", []string{
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "1",
			"2", "3", "4", "5", "6", "7", "8", "9"}},
		// FIXME
		//{"Héllo", []string{"H", "e", "́", "l", "l", "o"}},
		//{"𐘀𐘁𐘂", []string{"𐘀", "𐘁", "𐘂"}},
	}

	tokenizer := newQGram(1)
	check(t, tokenizer, cases)
}

func TestQGram2(t *testing.T) {
	cases := []tCase{
		{"", []string{}},
		{"1", []string{"1"}},
		{"12", []string{"12"}},
		{"123456789", []string{"12", "23", "34", "45", "56", "67", "78", "89"}},
		{"123456789123456789", []string{
			"12", "23", "34", "45", "56", "67", "78", "89",
			"91", "12", "23", "34", "45", "56", "67", "78",
			"89"}},
		// {"Héllo", []string{"He", "é", "́l", "ll", "lo"}},
		// {"𐘀𐘁𐘂", []string{"𐘀𐘁", "𐘁𐘂"}},
		// {"𐘀𐘁", []string{"𐘀𐘁"}},
	}

	tokenizer := newQGram(2)
	check(t, tokenizer, cases)
}

func TestQGram2WithPadding(t *testing.T) {
	cases := []tCase{
		{"", []string{}},
		{"1", []string{"@1", "1@"}},
		{"12", []string{"@1", "12", "2@"}},
	}

	tokenizer := newQGramExtended(2, "@", "@")
	check(t, tokenizer, cases)
}

func TestQGram2WithTwoSidedPadding(t *testing.T) {
	cases := []tCase{
		{"", []string{}},
		{"1", []string{"L1", "1R"}},
		{"12", []string{"L1", "12", "2R"}},
	}

	tokenizer := newQGramExtended(2, "L", "R")
	check(t, tokenizer, cases)
}

func TestQGram3(t *testing.T) {
	cases := []tCase{
		{"", []string{}},
		{"1", []string{"1"}},
		{"12", []string{"12"}},
		{"123", []string{"123"}},
		{"12345678", []string{"123", "234", "345", "456", "567", "678"}},
		{"123123", []string{"123", "231", "312", "123"}},
		// {"Héllo", []string{"Hé", "él", "́ll", "llo"}},
		// {"𐘀𐘁𐘂", []string{"𐘀𐘁𐘂"}},
	}

	tokenizer := newQGram(3)
	check(t, tokenizer, cases)
}

func TestQGram3WithPadding(t *testing.T) {
	cases := []tCase{
		{"", []string{}},
		{"1", []string{"##1", "#1#", "1##"}},
		{"12", []string{"##1", "#12", "12#", "2##"}},
		{"123", []string{"##1", "#12", "123", "23#", "3##"}},
		{"12345678", []string{"##1", "#12", "123", "234", "345", "456", "567", "678", "78#", "8##"}},
		{"123123", []string{"##1", "#12", "123", "231", "312", "123", "23#", "3##"}},
	}

	tokenizer := newQGramExtended(3, "#", "#")
	check(t, tokenizer, cases)
}

func check(t *testing.T, q tokenizer, cases []tCase) {
	for _, c := range cases {
		actual := q.Tokenize(c.Input)
		if len(actual) != len(c.Expected) {
			t.Fatalf("wrong length %d expected %d, input: %s", len(actual), len(c.Expected), c.Input)
		}
		for i, v := range actual {
			if v != c.Expected[i] {
				t.Fatalf("%s != %s, input: %s", v, c.Expected[i], c.Input)
			}
		}
	}
}

type mCase struct {
	expected float32
	a        string
	b        string
}

func TestQGramsMetric(t *testing.T) {
	cases := []mCase{
		{0.7857, "test string1", "test string2"},
		{0.4000, "test", "test string2"},
		{0.0000, "", "test string2"},
		{0.7059, "aaa bbb ccc ddd", "aaa bbb ccc eee"},
		{0.6667, "a b c d", "a b c e"},
	}

	d := qGramsDistance()
	for _, c := range cases {
		actual := d.Compare(c.a, c.b)
		if fmt.Sprintf("%.4f", actual) != fmt.Sprintf("%.4f", c.expected) {
			t.Fatalf("%.4f != %.4f for '%s' and '%s", actual, c.expected, c.a, c.b)
		}
	}
}
