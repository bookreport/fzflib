package fzflib

import (
	"github.com/bookreport/fzflib/algo"
	"github.com/bookreport/fzflib/util"
)

func Search(query string, content []string) []string {
	patternBuilder := func(runes []rune) *pattern {
		return buildPattern(
			true,
			algo.FuzzyMatchV2,
			true,
			CaseSmart,
			true,
			true,
			false,
			make([]Range, 0),
			Delimiter{},
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

	var results []string
	pattern := patternBuilder([]rune(query))
	slab := util.MakeSlab(slab16Size, slab32Size)
	for _, c := range content {
		var i item
		chunkList.trans(&i, []byte(c))
		if result, _, _ := pattern.MatchItem(&i, false, slab); result != nil {
			results = append(results, i.text.ToString())
		}
	}

	return results
}
