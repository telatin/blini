package sketching

import (
	"fmt"
	"math"

	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/maps"
)

// Index allows quick lookups for sketches.
type Index struct {
	idx map[uint64][]int
	mx  uint64
}

// NewIndex returns a new index that stores 1/scale of hashes.
func NewIndex(scale uint64) *Index {
	return &Index{
		idx: map[uint64][]int{},
		mx:  math.MaxUint64 / scale,
	}
}

// Add adds the given sketch with the given serial number.
func (idx *Index) Add(s []uint64, i int) {
	for _, x := range s {
		if x > idx.mx {
			break
		}
		idx.idx[x] = append(idx.idx[x], i)
	}
}

// Search returns serial numbers of sketches that share hashes with
// the given sketch.
func (idx *Index) Search(s []uint64) []int {
	set := sets.Set[int]{}
	for _, x := range s {
		if x > idx.mx {
			break
		}
		set.Add(idx.idx[x]...)
	}
	return maps.Keys(set)
}

// Clean removes keys with only one element.
// Use only for clustering.
func (idx *Index) Clean() {
	n1 := len(idx.idx)
	idx.idx = snm.FilterMap(idx.idx, func(k uint64, v []int) bool {
		return len(v) > 1
	})
	n2 := len(idx.idx)
	fmt.Printf("Cleaning: %d ==> %d (%.0f%%)\n",
		n1, n2, float64(n2)/float64(n1)*100)
}
