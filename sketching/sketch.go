package sketching

import (
	"bytes"
	"math"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/hashx"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/maps"
)

func Sketch(seq []byte, k int, scale uint64) []uint64 {
	seq = bytes.ToUpper(seq)
	hashes := make(sets.Set[uint64], len(seq)/int(scale))
	mx := math.MaxUint64 / scale
	for sseq := range sequtil.SubsequencesWith(seq, "atcgATCG") {
		for s := range sequtil.CanonicalSubsequences(sseq, k) {
			h := hashx.Bytes(s)
			if h > mx {
				continue
			}
			hashes.Add(h)
		}
	}
	return snm.Sorted(maps.Keys(hashes))
}
