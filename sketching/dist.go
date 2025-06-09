package sketching

import "github.com/fluhus/biostuff/mash"

const (
	// Punish containment for possible mismatches.
	compensatingCont = true
)

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

func Jaccard(a, b []uint64) float64 {
	i := common(a, b)
	u := len(a) + len(b) - i
	return float64(i) / float64(u)
}

func Containment(a, b []uint64) float64 {
	i := common(a, b)
	if compensatingCont {
		return float64(i) / float64(len(a)+(len(a)-i))
	} else {
		return float64(i) / float64(len(a))
	}
}

func MyDist(a, b []uint64, alen, blen int, k int) float64 {
	if alen > blen { // a should be the smaller.
		a, b = b, a
		alen, blen = blen, alen
	}
	r := float64(alen) / float64(blen)
	d := mash.FromJaccard(Containment(a, b), k)
	return d*r + (1 - r)
}
