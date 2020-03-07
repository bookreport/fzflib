package fzflib

import (
	"fmt"
	"log"
)

func logResult(r *result) {
	log.Println(fmt.Sprintf("{ item: { text: '%v', origText: '%v' }, points: %v }", r.item.text, r.item.origText, r.points))
}

func logResultSlice(results []result) {
	l := "[\n"
	log.Println("len(results)", len(results))
	for _, r := range results {
		l += fmt.Sprintf("{ item: { text: '%v' }, points: %v }", r.item.text.ToString(), r.points)
		l += ",\n"
	}

	l += "\n]"
	log.Println(l)
}
