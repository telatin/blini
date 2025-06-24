package main

import (
	"cmp"
	"slices"
	"testing"
)

func TestSortedPerm(t *testing.T) {
	input := []string{"a", "g", "d", "b"}
	want := []int{0, 3, 2, 1}
	got := sortedPerm(input, cmp.Compare)
	if !slices.Equal(got, want) {
		t.Fatalf("sortedPerm(%q)=%v, want %v", input, got, want)
	}
}
