// Sketch I/O logic.

package main

import (
	"fmt"
	"io"
	"iter"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/blini/sketching"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
)

// Holds sketches of input sequences and metadata.
type sketches struct {
	skch  [][]uint64 // Frac min hash sketches.
	lens  []int      // Sequence lengths.
	names []string   // Sequence names.
	scale uint64     // Kmer selection scale.
}

type sketchEntry struct {
	s     []uint64 // Sketch hashes.
	ln    int      // Sequence length.
	name  string   // Sequence name.
	scale uint64   // Kmer selection scale.
}

// Sketches an input fasta file and iterates over the sketches.
func sketchFile(file string) iter.Seq2[sketchEntry, error] {
	return func(yield func(sketchEntry, error) bool) {
		for fa, err := range fasta.File(file) {
			if err != nil {
				yield(sketchEntry{}, err)
				return
			}
			var e sketchEntry
			e.s = sketching.Sketch(fa.Sequence, kmerLen, *scale)
			e.ln = len(fa.Sequence)
			e.name = string(fa.Name)
			e.scale = *scale
			if !yield(e, nil) {
				return
			}
		}
	}
}

// Iterates over sketches in a file.
func readSketches(file string) iter.Seq2[sketchEntry, error] {
	return func(yield func(sketchEntry, error) bool) {
		f, err := aio.Open(file)
		if err != nil {
			yield(sketchEntry{}, err)
			return
		}
		defer f.Close()

		for {
			var e sketchEntry
			err := bnry.Read(f, &e.s, &e.ln, &e.name, &e.scale)
			if err != nil {
				if err == io.EOF {
					return
				}
				yield(sketchEntry{}, err)
				return
			}
			if !yield(e, nil) {
				return
			}
		}
	}
}

// Collects sketches from an iterator,
// validating that their scales are the same.
func collectSketches(seq iter.Seq2[sketchEntry, error]) (sketches, error) {
	skch := sketches{}
	first := true
	pt := ptimer.New()
	for s, err := range seq {
		if err != nil {
			return skch, err
		}
		if first {
			skch.scale = s.scale
			first = false
		} else {
			if s.scale != skch.scale {
				return skch, fmt.Errorf("mismatching scales: %d, %d",
					skch.scale, s.scale)
			}
		}
		skch.skch = append(skch.skch, s.s)
		skch.lens = append(skch.lens, s.ln)
		skch.names = append(skch.names, s.name)
		pt.Inc()
	}
	pt.Done()
	return skch, nil
}
