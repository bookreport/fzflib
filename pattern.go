package fzflib

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bookreport/fzflib/algo"
	"github.com/bookreport/fzflib/util"
)

// fuzzy
// 'exact
// ^prefix-exact
// suffix-exact$
// !inverse-exact
// !'inverse-fuzzy
// !^inverse-prefix-exact
// !inverse-suffix-exact$

type termType int

const (
	termFuzzy termType = iota
	termExact
	termPrefix
	termSuffix
	termEqual
)

type term struct {
	typ           termType
	inv           bool
	text          []rune
	caseSensitive bool
}

// String returns the string representation of a term.
func (t term) String() string {
	return fmt.Sprintf("term{typ: %d, inv: %v, text: []rune(%q), caseSensitive: %v}", t.typ, t.inv, string(t.text), t.caseSensitive)
}

type termSet []term

// pattern represents search pattern
type pattern struct {
	fuzzy         bool
	fuzzyAlgo     algo.Algo
	extended      bool
	caseSensitive bool
	normalize     bool
	forward       bool
	text          []rune
	termSets      []termSet
	sortable      bool
	cacheable     bool
	cacheKey      string
	delimiter     inputDelimiter
	nth           []exprRange
	procFun       map[termType]algo.Algo
}

var (
	_patternCache map[string]*pattern
	_splitRegex   *regexp.Regexp
	_cache        chunkCache
)

func init() {
	_splitRegex = regexp.MustCompile(" +")
	clearPatternCache()
	clearChunkCache()
}

func clearPatternCache() {
	// We can uniquely identify the pattern for a given string since
	// search mode and caseMode do not change while the program is running
	_patternCache = make(map[string]*pattern)
}

func clearChunkCache() {
	_cache = newChunkCache()
}

// buildPattern builds pattern object from the given arguments
func buildPattern(
	fuzzy bool,
	fuzzyAlgo algo.Algo,
	extended bool,
	caseMode searchCase,
	normalize bool,
	forward bool,
	cacheable bool,
	nth []exprRange,
	delimiter inputDelimiter,
	runes []rune,
) *pattern {

	var asString string
	if extended {
		asString = strings.TrimLeft(string(runes), " ")
		for strings.HasSuffix(asString, " ") && !strings.HasSuffix(asString, "\\ ") {
			asString = asString[:len(asString)-1]
		}
	} else {
		asString = string(runes)
	}

	cached, found := _patternCache[asString]
	if found {
		return cached
	}

	caseSensitive := true
	sortable := true
	termSets := []termSet{}

	if extended {
		termSets = parseTerms(fuzzy, caseMode, normalize, asString)
		// We should not sort the result if there are only inverse search terms
		sortable = false
	Loop:
		for _, termSet := range termSets {
			for idx, term := range termSet {
				if !term.inv {
					sortable = true
				}
				// If the query contains inverse search terms or OR operators,
				// we cannot cache the search scope
				if !cacheable || idx > 0 || term.inv || fuzzy && term.typ != termFuzzy || !fuzzy && term.typ != termExact {
					cacheable = false
					if sortable {
						// Can't break until we see at least one non-inverse term
						break Loop
					}
				}
			}
		}
	} else {
		lowerString := strings.ToLower(asString)
		caseSensitive = caseMode == searchCaseRespect ||
			caseMode == searchCaseSmart && lowerString != asString
		if !caseSensitive {
			asString = lowerString
		}
	}

	ptr := &pattern{
		fuzzy:         fuzzy,
		fuzzyAlgo:     fuzzyAlgo,
		extended:      extended,
		caseSensitive: caseSensitive,
		normalize:     normalize,
		forward:       forward,
		text:          []rune(asString),
		termSets:      termSets,
		sortable:      sortable,
		cacheable:     cacheable,
		nth:           nth,
		delimiter:     delimiter,
		procFun:       make(map[termType]algo.Algo)}

	ptr.cacheKey = ptr.buildCacheKey()
	ptr.procFun[termFuzzy] = fuzzyAlgo
	ptr.procFun[termEqual] = algo.EqualMatch
	ptr.procFun[termExact] = algo.ExactMatchNaive
	ptr.procFun[termPrefix] = algo.PrefixMatch
	ptr.procFun[termSuffix] = algo.SuffixMatch

	_patternCache[asString] = ptr
	return ptr
}

func parseTerms(fuzzy bool, caseMode searchCase, normalize bool, str string) []termSet {
	str = strings.Replace(str, "\\ ", "\t", -1)
	tokens := _splitRegex.Split(str, -1)
	sets := []termSet{}
	set := termSet{}
	switchSet := false
	afterBar := false
	for _, token := range tokens {
		typ, inv, text := termFuzzy, false, strings.Replace(token, "\t", " ", -1)
		lowerText := strings.ToLower(text)
		caseSensitive := caseMode == searchCaseRespect ||
			caseMode == searchCaseSmart && text != lowerText
		if !caseSensitive {
			text = lowerText
		}
		if !fuzzy {
			typ = termExact
		}

		if len(set) > 0 && !afterBar && text == "|" {
			switchSet = false
			afterBar = true
			continue
		}
		afterBar = false

		if strings.HasPrefix(text, "!") {
			inv = true
			typ = termExact
			text = text[1:]
		}

		if text != "$" && strings.HasSuffix(text, "$") {
			typ = termSuffix
			text = text[:len(text)-1]
		}

		if strings.HasPrefix(text, "'") {
			// Flip exactness
			if fuzzy && !inv {
				typ = termExact
				text = text[1:]
			} else {
				typ = termFuzzy
				text = text[1:]
			}
		} else if strings.HasPrefix(text, "^") {
			if typ == termSuffix {
				typ = termEqual
			} else {
				typ = termPrefix
			}
			text = text[1:]
		}

		if len(text) > 0 {
			if switchSet {
				sets = append(sets, set)
				set = termSet{}
			}
			textRunes := []rune(text)
			if normalize {
				textRunes = algo.NormalizeRunes(textRunes)
			}
			set = append(set, term{
				typ:           typ,
				inv:           inv,
				text:          textRunes,
				caseSensitive: caseSensitive})
			switchSet = true
		}
	}
	if len(set) > 0 {
		sets = append(sets, set)
	}
	return sets
}

// IsEmpty returns true if the pattern is effectively empty
func (p *pattern) IsEmpty() bool {
	if !p.extended {
		return len(p.text) == 0
	}
	return len(p.termSets) == 0
}

// AsString returns the search query in string type
func (p *pattern) AsString() string {
	return string(p.text)
}

func (p *pattern) buildCacheKey() string {
	if !p.extended {
		return p.AsString()
	}
	cacheableTerms := []string{}
	for _, termSet := range p.termSets {
		if len(termSet) == 1 && !termSet[0].inv && (p.fuzzy || termSet[0].typ == termExact) {
			cacheableTerms = append(cacheableTerms, string(termSet[0].text))
		}
	}
	return strings.Join(cacheableTerms, "\t")
}

// CacheKey is used to build string to be used as the key of result cache
func (p *pattern) CacheKey() string {
	return p.cacheKey
}

// Match returns the list of matches Items in the given chunk
func (p *pattern) Match(chunk *chunk, slab *util.Slab) []result {
	// chunkCache: Exact match
	cacheKey := p.CacheKey()
	if p.cacheable {
		if cached := _cache.Lookup(chunk, cacheKey); cached != nil {
			return cached
		}
	}

	// Prefix/suffix cache
	space := _cache.Search(chunk, cacheKey)

	matches := p.matchChunk(chunk, space, slab)

	if p.cacheable {
		_cache.Add(chunk, cacheKey, matches)
	}
	return matches
}

func (p *pattern) matchChunk(chunk *chunk, space []result, slab *util.Slab) []result {
	matches := []result{}

	if space == nil {
		for idx := 0; idx < chunk.count; idx++ {
			if match, _, _ := p.MatchItem(&chunk.items[idx], false, slab); match != nil {
				matches = append(matches, *match)
			}
		}
	} else {
		for _, result := range space {
			if match, _, _ := p.MatchItem(result.item, false, slab); match != nil {
				matches = append(matches, *match)
			}
		}
	}
	return matches
}

// MatchItem returns true if the item is a match
func (p *pattern) MatchItem(item *item, withPos bool, slab *util.Slab) (*result, []substrOffset, *[]int) {
	if p.extended {
		if offsets, bonus, pos := p.extendedMatch(item, withPos, slab); len(offsets) == len(p.termSets) {
			result := buildResult(item, offsets, bonus)
			return &result, offsets, pos
		}
		return nil, nil, nil
	}
	offset, bonus, pos := p.basicMatch(item, withPos, slab)
	if sidx := offset[0]; sidx >= 0 {
		offsets := []substrOffset{offset}
		result := buildResult(item, offsets, bonus)
		return &result, offsets, pos
	}
	return nil, nil, nil
}

func (p *pattern) basicMatch(item *item, withPos bool, slab *util.Slab) (substrOffset, int, *[]int) {
	var input []token
	if len(p.nth) == 0 {
		input = []token{token{text: &item.text, prefixLength: 0}}
	} else {
		input = p.transformInput(item)
	}
	if p.fuzzy {
		return p.iter(p.fuzzyAlgo, input, p.caseSensitive, p.normalize, p.forward, p.text, withPos, slab)
	}
	return p.iter(algo.ExactMatchNaive, input, p.caseSensitive, p.normalize, p.forward, p.text, withPos, slab)
}

func (p *pattern) extendedMatch(item *item, withPos bool, slab *util.Slab) ([]substrOffset, int, *[]int) {
	var input []token
	if len(p.nth) == 0 {
		input = []token{token{text: &item.text, prefixLength: 0}}
	} else {
		input = p.transformInput(item)
	}
	offsets := []substrOffset{}
	var totalScore int
	var allPos *[]int
	if withPos {
		allPos = &[]int{}
	}
	for _, termSet := range p.termSets {
		var offset substrOffset
		var currentScore int
		matched := false
		for _, term := range termSet {
			pfun := p.procFun[term.typ]
			off, score, pos := p.iter(pfun, input, term.caseSensitive, p.normalize, p.forward, term.text, withPos, slab)
			if sidx := off[0]; sidx >= 0 {
				if term.inv {
					continue
				}
				offset, currentScore = off, score
				matched = true
				if withPos {
					if pos != nil {
						*allPos = append(*allPos, *pos...)
					} else {
						for idx := off[0]; idx < off[1]; idx++ {
							*allPos = append(*allPos, int(idx))
						}
					}
				}
				break
			} else if term.inv {
				offset, currentScore = substrOffset{0, 0}, 0
				matched = true
				continue
			}
		}
		if matched {
			offsets = append(offsets, offset)
			totalScore += currentScore
		}
	}
	return offsets, totalScore, allPos
}

func (p *pattern) transformInput(item *item) []token {
	if item.transformed != nil {
		return *item.transformed
	}

	tokens := tokenize(item.text.ToString(), p.delimiter)
	ret := transform(tokens, p.nth)
	item.transformed = &ret
	return ret
}

func (p *pattern) iter(pfun algo.Algo, tokens []token, caseSensitive bool, normalize bool, forward bool, pattern []rune, withPos bool, slab *util.Slab) (substrOffset, int, *[]int) {
	for _, part := range tokens {
		if res, pos := pfun(caseSensitive, normalize, forward, part.text, pattern, withPos, slab); res.Start >= 0 {
			sidx := int32(res.Start) + part.prefixLength
			eidx := int32(res.End) + part.prefixLength
			if pos != nil {
				for idx := range *pos {
					(*pos)[idx] += int(part.prefixLength)
				}
			}
			return substrOffset{sidx, eidx}, res.Score, pos
		}
	}
	return substrOffset{-1, -1}, 0, nil
}
