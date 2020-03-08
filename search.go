package fzflib

import (
	"sort"

	"github.com/bookreport/fzflib/algo"
	"github.com/bookreport/fzflib/util"
)

func Search(query string, content []string) []string {
	sortCriteria = []criterion{byScore, byLength}
	patternBuilder := func(runes []rune) *pattern {
		return buildPattern(
			true,
			algo.FuzzyMatchV2,
			true,
			searchCaseSmart,
			true,
			true,
			false,
			make([]exprRange, 0),
			inputDelimiter{},
			runes,
		)
	}

	var itemIndex int32
	chunkList := newChunkList(func(item *item, data []byte) bool {
		item.text = util.ToChars(data)
		item.text.Index = itemIndex
		itemIndex++
		return true
	})

	var results []result
	pattern := patternBuilder([]rune(query))
	slab := util.MakeSlab(slab16Size, slab32Size)
	for _, c := range content {
		var i item
		chunkList.trans(&i, []byte(c))
		if result, _, _ := pattern.MatchItem(&i, false, slab); result != nil {
			results = append(results, *result)
		}
	}

	sort.Sort(byRelevance(results))

	var resultStrings []string
	for _, r := range results {
		resultStrings = append(resultStrings, r.item.text.ToString())
	}

	return resultStrings
}
