// Generates test data for Blini's CI tests.
package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/blini/paper/simul"
	"github.com/fluhus/gostuff/aio"
)

const (
	outDir = "testdata"
)

func main() {
	if err := genClustData(); err != nil {
		panic(err)
	}
	if err := genSearchData(); err != nil {
		panic(err)
	}
}

func genClustData() error {
	const (
		nClusters = 3
		nElements = 3
		minLen    = 5000
		maxLen    = 10000
	)

	var fas []*fasta.Fasta
	for i := range nClusters {
		// Create cluster center.
		fa := &fasta.Fasta{
			Name:     fmt.Appendf(nil, "ref%d", i+1),
			Sequence: simul.RandSeq(minLen + rand.IntN(maxLen-minLen+1)),
		}
		fas = append(fas, fa)

		// Create elements.
		for j := range nElements {
			ssLen := len(fa.Sequence)/2 + rand.IntN(len(fa.Sequence)/2) - 1
			ss := &fasta.Fasta{
				Name:     fmt.Appendf(nil, "ref%d.%d", i+1, j+1),
				Sequence: simul.RandSubseq(fa.Sequence, ssLen),
			}
			ss.Sequence = simul.MutSeqPerc(ss.Sequence, 1)
			fas = append(fas, ss)
		}
	}

	shuffle(fas)

	fout, err := aio.Create(outDir + "/clust.fa.zst")
	if err != nil {
		return err
	}
	defer fout.Close()
	for _, fa := range fas {
		if err := fa.Write(fout); err != nil {
			return nil
		}
	}
	return nil
}

func genSearchData() error {
	const (
		nRefs    = 3
		nQueries = 3
		minLen   = 5000
		maxLen   = 10000
	)

	var refs []*fasta.Fasta
	var queries []*fasta.Fasta
	for i := range nRefs {
		// Create reference sequence.
		fa := &fasta.Fasta{
			Name:     fmt.Appendf(nil, "ref%d", i+1),
			Sequence: simul.RandSeq(minLen + rand.IntN(maxLen-minLen+1)),
		}
		refs = append(refs, fa)

		// Create elements.
		for j := range nQueries {
			ssLen := len(fa.Sequence)/2 + rand.IntN(len(fa.Sequence)/2) - 1
			ss := &fasta.Fasta{
				Name:     fmt.Appendf(nil, "query%d.%d", i+1, j+1),
				Sequence: simul.RandSubseq(fa.Sequence, ssLen),
			}
			ss.Sequence = simul.MutSeqPerc(ss.Sequence, 2)
			queries = append(queries, ss)
		}
	}

	shuffle(refs)
	shuffle(queries)

	rout, err := aio.Create(outDir + "/refs.fa.zst")
	if err != nil {
		return err
	}
	defer rout.Close()
	for _, fa := range refs {
		if err := fa.Write(rout); err != nil {
			return nil
		}
	}

	qout, err := aio.Create(outDir + "/queries.fa.zst")
	if err != nil {
		return err
	}
	defer qout.Close()
	for _, fa := range queries {
		if err := fa.Write(qout); err != nil {
			return nil
		}
	}
	return nil
}

func shuffle[T any](a []T) {
	rand.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
}
