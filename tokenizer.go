package fzflib

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/bookreport/fzflib/util"
)

// searchCase denotes case-sensitivity of search
type searchCase int

// searchCase-sensitivities
const (
	searchCaseSmart searchCase = iota
	searchCaseIgnore
	searchCaseRespect
)

const rangeEllipsis = 0

// exprRange represents nth-expression
type exprRange struct {
	begin int
	end   int
}

// token contains the tokenized part of the strings and its prefix length
type token struct {
	text         *util.Chars
	prefixLength int32
}

// String returns the string representation of a token.
func (t token) String() string {
	return fmt.Sprintf("token{text: %s, prefixLength: %d}", t.text, t.prefixLength)
}

// inputDelimiter for tokenizing the input
type inputDelimiter struct {
	regex *regexp.Regexp
	str   *string
}

// String returns the string representation of a inputDelimiter.
func (d inputDelimiter) String() string {
	return fmt.Sprintf("inputDelimiter{regex: %v, str: &%q}", d.regex, *d.str)
}

func withPrefixLengths(tokens []string, begin int) []token {
	ret := make([]token, len(tokens))

	prefixLength := begin
	for idx := range tokens {
		chars := util.ToChars([]byte(tokens[idx]))
		ret[idx] = token{&chars, int32(prefixLength)}
		prefixLength += chars.Length()
	}
	return ret
}

const (
	awkNil = iota
	awkBlack
	awkWhite
)

func awkTokenizer(input string) ([]string, int) {
	// 9, 32
	ret := []string{}
	prefixLength := 0
	state := awkNil
	begin := 0
	end := 0
	for idx := 0; idx < len(input); idx++ {
		r := input[idx]
		white := r == 9 || r == 32
		switch state {
		case awkNil:
			if white {
				prefixLength++
			} else {
				state, begin, end = awkBlack, idx, idx+1
			}
		case awkBlack:
			end = idx + 1
			if white {
				state = awkWhite
			}
		case awkWhite:
			if white {
				end = idx + 1
			} else {
				ret = append(ret, input[begin:end])
				state, begin, end = awkBlack, idx, idx+1
			}
		}
	}
	if begin < end {
		ret = append(ret, input[begin:end])
	}
	return ret, prefixLength
}

// tokenize splits apart the given string using the delimiter
func tokenize(text string, delimiter inputDelimiter) []token {
	if delimiter.str == nil && delimiter.regex == nil {
		// AWK-style (\S+\s*)
		tokens, prefixLength := awkTokenizer(text)
		return withPrefixLengths(tokens, prefixLength)
	}

	if delimiter.str != nil {
		return withPrefixLengths(strings.SplitAfter(text, *delimiter.str), 0)
	}

	// FIXME performance
	var tokens []string
	if delimiter.regex != nil {
		for len(text) > 0 {
			loc := delimiter.regex.FindStringIndex(text)
			if len(loc) < 2 {
				loc = []int{0, len(text)}
			}
			last := util.Max(loc[1], 1)
			tokens = append(tokens, text[:last])
			text = text[last:]
		}
	}
	return withPrefixLengths(tokens, 0)
}

func joinTokens(tokens []token) string {
	var output bytes.Buffer
	for _, token := range tokens {
		output.WriteString(token.text.ToString())
	}
	return output.String()
}

// transform is used to transform the input when --with-nth option is given
func transform(tokens []token, withNth []exprRange) []token {
	transTokens := make([]token, len(withNth))
	numTokens := len(tokens)
	for idx, r := range withNth {
		parts := []*util.Chars{}
		minIdx := 0
		if r.begin == r.end {
			idx := r.begin
			if idx == rangeEllipsis {
				chars := util.ToChars([]byte(joinTokens(tokens)))
				parts = append(parts, &chars)
			} else {
				if idx < 0 {
					idx += numTokens + 1
				}
				if idx >= 1 && idx <= numTokens {
					minIdx = idx - 1
					parts = append(parts, tokens[idx-1].text)
				}
			}
		} else {
			var begin, end int
			if r.begin == rangeEllipsis { // ..N
				begin, end = 1, r.end
				if end < 0 {
					end += numTokens + 1
				}
			} else if r.end == rangeEllipsis { // N..
				begin, end = r.begin, numTokens
				if begin < 0 {
					begin += numTokens + 1
				}
			} else {
				begin, end = r.begin, r.end
				if begin < 0 {
					begin += numTokens + 1
				}
				if end < 0 {
					end += numTokens + 1
				}
			}
			minIdx = util.Max(0, begin-1)
			for idx := begin; idx <= end; idx++ {
				if idx >= 1 && idx <= numTokens {
					parts = append(parts, tokens[idx-1].text)
				}
			}
		}
		// Merge multiple parts
		var merged util.Chars
		switch len(parts) {
		case 0:
			merged = util.ToChars([]byte{})
		case 1:
			merged = *parts[0]
		default:
			var output bytes.Buffer
			for _, part := range parts {
				output.WriteString(part.ToString())
			}
			merged = util.ToChars(output.Bytes())
		}

		var prefixLength int32
		if minIdx < numTokens {
			prefixLength = tokens[minIdx].prefixLength
		} else {
			prefixLength = 0
		}
		transTokens[idx] = token{&merged, prefixLength}
	}
	return transTokens
}
