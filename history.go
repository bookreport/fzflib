package fzflib

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

// history struct represents input history
type history struct {
	path     string
	lines    []string
	modified map[int]string
	maxSize  int
	cursor   int
}

// newHistory returns the pointer to a new history struct
func newHistory(path string, maxSize int) (*history, error) {
	fmtError := func(e error) error {
		if os.IsPermission(e) {
			return errors.New("permission denied: " + path)
		}
		return errors.New("invalid history file: " + e.Error())
	}

	// Read history file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		// If it doesn't exist, check if we can create a file with the name
		if os.IsNotExist(err) {
			data = []byte{}
			if err := ioutil.WriteFile(path, data, 0600); err != nil {
				return nil, fmtError(err)
			}
		} else {
			return nil, fmtError(err)
		}
	}
	// Split lines and limit the maximum number of lines
	lines := strings.Split(strings.Trim(string(data), "\n"), "\n")
	if len(lines[len(lines)-1]) > 0 {
		lines = append(lines, "")
	}
	return &history{
		path:     path,
		maxSize:  maxSize,
		lines:    lines,
		modified: make(map[int]string),
		cursor:   len(lines) - 1}, nil
}

func (h *history) append(line string) error {
	// We don't append empty lines
	if len(line) == 0 {
		return nil
	}

	lines := append(h.lines[:len(h.lines)-1], line)
	if len(lines) > h.maxSize {
		lines = lines[len(lines)-h.maxSize:]
	}
	h.lines = append(lines, "")
	return ioutil.WriteFile(h.path, []byte(strings.Join(h.lines, "\n")), 0600)
}

func (h *history) override(str string) {
	// You can update the history but they're not written to the file
	if h.cursor == len(h.lines)-1 {
		h.lines[h.cursor] = str
	} else if h.cursor < len(h.lines)-1 {
		h.modified[h.cursor] = str
	}
}

func (h *history) current() string {
	if str, prs := h.modified[h.cursor]; prs {
		return str
	}
	return h.lines[h.cursor]
}

func (h *history) previous() string {
	if h.cursor > 0 {
		h.cursor--
	}
	return h.current()
}

func (h *history) next() string {
	if h.cursor < len(h.lines)-1 {
		h.cursor++
	}
	return h.current()
}
