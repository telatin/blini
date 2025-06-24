// Generates a simulated dataset for clustering.
package main

import (
	"fmt"
	"iter"
	"math"
	"math/rand/v2"
	"os"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/blini/publication/simul"
	"github.com/fluhus/gostuff/aio"
)

const (
	testDataDir = "publication/testdata"
)

func main() {
	fmt.Println("Generating SNP dataset")
	if err := generateSNPs(); err != nil {
		panic(err)
	}
	fmt.Println("Generating fragment dataset")
	if err := generateFragments(); err != nil {
		panic(err)
	}
}

// Generates simulated sequences with SNPs.
func generateSNPs() error {
	if err := os.MkdirAll(testDataDir, 0o755); err != nil {
		return err
	}
	f, err := aio.Create(testDataDir + "/snps.fa")
	if err != nil {
		return err
	}
	defer f.Close()

	i := 0
	for ln := range expRange(10000, 1000000, 300) {
		i++
		seq := simul.RandSeq(ln)
		fa := &fasta.Fasta{
			Name:     fmt.Append(nil, i, ".1"),
			Sequence: seq,
		}
		if err := fa.Write(f); err != nil {
			return err
		}

		for j := range 2 {
			mut := simul.MutSeqPerc(seq, 4)
			if rand.IntN(2) == 0 {
				mut = sequtil.ReverseComplement(nil, mut)
			}
			fa = &fasta.Fasta{
				Name:     fmt.Append(nil, i, ".", j+2),
				Sequence: mut,
			}
			if err := fa.Write(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// Generates simulated sequences with fragments.
func generateFragments() error {
	if err := os.MkdirAll(testDataDir, 0o755); err != nil {
		return err
	}
	f, err := aio.Create(testDataDir + "/frag.fa")
	if err != nil {
		return err
	}
	defer f.Close()

	i := 0
	for ln := range expRange(10000, 1000000, 300) {
		i++
		seq := simul.RandSeq(ln)
		fa := &fasta.Fasta{
			Name:     fmt.Append(nil, i, ".1"),
			Sequence: seq,
		}
		if err := fa.Write(f); err != nil {
			return err
		}

		for j, ratio := range []int{2, 5, 10, 20, 50, 100} {
			mut := simul.RandSubseq(seq, len(seq)/ratio)
			if len(mut) < 1000 {
				continue
			}
			// Throw in some snps for good measure.
			mut = simul.MutSeqPerc(mut, 1)
			if rand.IntN(2) == 0 {
				mut = sequtil.ReverseComplement(nil, mut)
			}
			fa = &fasta.Fasta{
				Name:     fmt.Append(nil, i, ".", j+2),
				Sequence: mut,
			}
			if err := fa.Write(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// A range with exponential steps rather than linear.
func expRange(from, to, steps int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := range steps {
			n := float64(from) * math.Pow(float64(to)/float64(from),
				float64(i)/float64(steps-1))
			if !yield(int(math.Round(n))) {
				break
			}
		}
	}
}
