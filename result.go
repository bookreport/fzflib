package fzflib

import (
	"math"
	"sort"
	"unicode"

	"github.com/bookreport/fzflib/util"
)

// Sort criteria
type criterion int

const (
	byScore criterion = iota
	byLength
	byBegin
	byEnd
)

// substrOffset holds two 32-bit integers denoting the offsets of a matched substring
type substrOffset [2]int32

type result struct {
	item   *item
	points [4]uint16
}

func buildResult(item *item, offsets []substrOffset, score int) result {
	if len(offsets) > 1 {
		sort.Sort(byOrder(offsets))
	}

	result := result{item: item}
	numChars := item.text.Length()
	minBegin := math.MaxUint16
	minEnd := math.MaxUint16
	maxEnd := 0
	validOffsetFound := false
	for _, offset := range offsets {
		b, e := int(offset[0]), int(offset[1])
		if b < e {
			minBegin = util.Min(b, minBegin)
			minEnd = util.Min(e, minEnd)
			maxEnd = util.Max(e, maxEnd)
			validOffsetFound = true
		}
	}

	for idx, criterion := range sortCriteria {
		val := uint16(math.MaxUint16)
		switch criterion {
		case byScore:
			// Higher is better
			val = math.MaxUint16 - util.AsUint16(score)
		case byLength:
			val = item.TrimLength()
		case byBegin, byEnd:
			if validOffsetFound {
				whitePrefixLen := 0
				for idx := 0; idx < numChars; idx++ {
					r := item.text.Get(idx)
					whitePrefixLen = idx
					if idx == minBegin || !unicode.IsSpace(r) {
						break
					}
				}
				if criterion == byBegin {
					val = util.AsUint16(minEnd - whitePrefixLen)
				} else {
					val = util.AsUint16(math.MaxUint16 - math.MaxUint16*(maxEnd-whitePrefixLen)/int(item.TrimLength()))
				}
			}
		}
		result.points[3-idx] = val
	}

	return result
}

// Sort criteria to use. Never changes once fzf is started.
var sortCriteria []criterion

// Index returns ordinal index of the item
func (result *result) Index() int32 {
	return result.item.Index()
}

func minRank() result {
	return result{item: &minItem, points: [4]uint16{math.MaxUint16, 0, 0, 0}}
}

// byOrder is for sorting substring offsets
type byOrder []substrOffset

func (a byOrder) Len() int {
	return len(a)
}

func (a byOrder) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byOrder) Less(i, j int) bool {
	ioff := a[i]
	joff := a[j]
	return (ioff[0] < joff[0]) || (ioff[0] == joff[0]) && (ioff[1] <= joff[1])
}

// byRelevance is for sorting Items
type byRelevance []result

func (a byRelevance) Len() int {
	return len(a)
}

func (a byRelevance) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byRelevance) Less(i, j int) bool {
	return compareRanks(a[i], a[j], false)
}

// byRelevanceTac is for sorting Items
type byRelevanceTac []result

func (a byRelevanceTac) Len() int {
	return len(a)
}

func (a byRelevanceTac) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byRelevanceTac) Less(i, j int) bool {
	return compareRanks(a[i], a[j], true)
}
