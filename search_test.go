package fzflib

import (
	"testing"
)

func TestSearchOne(t *testing.T) {
	content := []string{
		"When nobody is around, the trees gossip about the people who have walked under them",
		"She had the gift of being able to paint songs.",
		"He was willing to find the depths of the rabbit hole in order to be with her.",
		"The tortoise jumped into the lake with dreams of becoming a sea turtle.",
		"It didn't make sense unless you had the power to eat colors.",
		"Sometimes I stare at a door or a wall and I wonder what is this reality, why am I alive, and what is this all about?",
		"The book is in front of the table.",
		"Her daily goal was to improve on yesterday.",
		"It was at that moment that he learned there are certain parts of the body that you should never Nair.",
		"There are few things better in life than a slice of pie.",
	}

	result := Search("daily", content)
	for _, r := range result {
		if r == "Her daily goal was to improve on yesterday." {
			return
		}
	}

	t.Error("expected to find 'Her daily goal was to improve on yesterday.' in result set")
}

func TestSearchTwo(t *testing.T) {
	content := []string{
		"When nobody is around, the trees gossip about the people who have walked under them",
		"She had the gift of being able to paint songs.",
		"He was willing to find the depths of the rabbit hole in order to be with her.",
		"The tortoise jumped into the lake with dreams of becoming a sea turtle.",
		"It didn't make sense unless you had the power to eat colors.",
		"Sometimes I stare at a door or a wall and I wonder what is this reality, why am I alive, and what is this all about?",
		"The book is in front of the table.",
		"Her daily goal was to improve on yesterday.",
		"It was at that moment that he learned there are certain parts of the body that you should never Nair.",
		"There are few things better in life than a slice of pie.",
	}

	result := Search("about", content)
	for _, r := range result {
		if r == "When nobody is around, the trees gossip about the people who have walked under them" {
			return
		}
	}

	t.Error("expected to find 'When nobody is around, the trees gossip about the people who have walked under them' in result set")
}
