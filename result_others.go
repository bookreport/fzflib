// +build !386,!amd64

package fzflib

func compareRanks(irank result, jrank result, tac bool) bool {
	for idx := 3; idx >= 0; idx-- {
		left := points[idx]
		right := points[idx]
		if left < right {
			return true
		} else if left > right {
			return false
		}
	}
	return (Index() <= Index()) != tac
}
