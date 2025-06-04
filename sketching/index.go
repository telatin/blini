package sketching

import (
	"fmt"
	"math"

	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/maps"
)

type Index struct {
	idx map[uint64][]int
	mx  uint64
}

func NewIndex(scale uint64) *Index {
	return &Index{
		idx: map[uint64][]int{},
		mx:  math.MaxUint64 / scale,
	}
}

func (idx *Index) Add(s []uint64, i int) {
	for _, x := range s {
		if x > idx.mx {
			break
		}
		idx.idx[x] = append(idx.idx[x], i)
	}
}

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

func (idx *Index) Clean() {
	n1 := len(idx.idx)
	idx.idx = snm.FilterMap(idx.idx, func(k uint64, v []int) bool {
		return len(v) > 1
	})
	n2 := len(idx.idx)
	fmt.Printf("Cleaning: %d ==> %d (%.0f%%)\n",
		n1, n2, float64(n2)/float64(n1)*100)
}
