// Package sketching provides sequence sketching and
// distance calculation functionality.
package sketching

import "github.com/fluhus/biostuff/mash"

const (
	// Punish containment for possible mismatches.
	compensatingCont = true

	// Add a ghost count to the union part of Jaccard calculations.
	// This "punishes" low kmer counts and makes them look less similar.
	ghostUnion = 0
)

// Counts the common elements in a and b, assuming they are both sorted.
func common(a, b []uint64) int {
	c, i, j := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if b[j] < a[i] {
			j++
		} else {
			c++
			i++
			j++
		}
	}
	return c
}

// Jaccard returns the Jaccard similarity between a and b.
func Jaccard(a, b []uint64) float64 {
	i := common(a, b)
	u := len(a) + len(b) - i
	u += ghostUnion
	return float64(i) / float64(u)
}

// Containment returns a Jaccard-like similarity for the containment
// of a in b.
func Containment(a, b []uint64) float64 {
	i := common(a, b)
	u := len(a)
	if compensatingCont {
		u += len(a) - i
	}
	u += ghostUnion
	return float64(i) / float64(u)
}

// MyDist returns a Mash distance with compensation for length
// difference.
func MyDist(a, b []uint64, alen, blen int, k int) float64 {
	if alen > blen { // a should be the smaller.
		a, b = b, a
		alen, blen = blen, alen
	}
	r := float64(alen) / float64(blen)
	d := mash.FromJaccard(Containment(a, b), k)
	return d*r + (1 - r)
}
