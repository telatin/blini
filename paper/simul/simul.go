// Package simul provides sequence simulation functions.
package simul

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/snm"
)

// RandSeq returns a random sequence of ATCG.
func RandSeq(n int) []byte {
	return snm.Slice(n, func(i int) byte {
		return sequtil.Iton(rand.N(4))
	})
}

// MutNuc returns a random nucleotide that's not a.
func MutNuc(a byte) byte {
	return sequtil.Iton((sequtil.Ntoi(a) + 1 + rand.IntN(3)) % 4)
}

// MutSeq returns a clone of seq with n nucleotides randomly changed.
func MutSeq(seq []byte, n int) []byte {
	seq = slices.Clone(seq)
	for _, i := range rand.Perm(len(seq))[:n] {
		seq[i] = MutNuc(seq[i])
	}
	return seq
}

// MutSeqPerc returns a clone of seq with perc percent of
// its nucleotides randomly changed.
func MutSeqPerc(seq []byte, perc int) []byte {
	n := len(seq) * perc / 100
	return MutSeq(seq, n)
}

// RandSubseq returns a subsequence of length n,
// chosen randomly from all n-long subsequences of seq.
func RandSubseq(seq []byte, n int) []byte {
	if n > len(seq) {
		panic(fmt.Sprintf("bad subsequence length: %d, want <=%d",
			n, len(seq)))
	}
	nss := len(seq) - n + 1
	i := rand.IntN(nss)
	return seq[i : i+n]
}
