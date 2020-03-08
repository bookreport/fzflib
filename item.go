package fzflib

import (
	"github.com/bookreport/fzflib/util"
)

// item represents each input line. 56 bytes.
type item struct {
	text        util.Chars // 32 = 24 + 1 + 1 + 2 + 4
	transformed *[]token   // 8
	origText    *[]byte    // 8
}

// Index returns ordinal index of the item
func (item *item) Index() int32 {
	return item.text.Index
}

var minItem = item{text: util.Chars{Index: -1}}

func (item *item) TrimLength() uint16 {
	return item.text.TrimLength()
}

// AsString returns the original string
func (item *item) AsString() string {
	return item.text.ToString()
}
