// Search logic.

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/biostuff/mash/v2"
	"github.com/fluhus/blini/sketching"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ptimer"
)

// Main function for search operation.
func mainSearch() error {
	fmt.Println("----------------")
	fmt.Println("SEARCH OPERATION")
	fmt.Println("----------------")
	var sk sketches
	var err error
	if strings.HasSuffix(*rFile, indexSuffix) {
		fmt.Println("Reading prepared sketches")
		sk, err = collectSketches(readSketches(*rFile))
	} else {
		fmt.Println("Sketching reference sequences")
		sk, err = collectSketches(sketchFile(*rFile))
	}
	if err != nil {
		return err
	}
	fmt.Println("Scale:", sk.scale)
	fmt.Println("Min sim:", *minSim)

	fmt.Println("Indexing")
	pt := ptimer.New()
	idx := sketching.NewIndex(sk.scale * idxScale)
	for i, s := range sk.skch {
		idx.Add(s, i)
		pt.Inc()
	}
	pt.Done()

	fmt.Println("Searching")
	var fout io.Writer
	if *oFile != "" {
		f, err := aio.Create(*oFile)
		if err != nil {
			return err
		}
		defer f.Close()
		fout = f
	} else {
		fout = io.Discard
	}
	out := csv.NewWriter(fout)
	defer out.Flush()

	out.Write([]string{"similarity", "query", "reference"})

	var matches int
	pt = ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d (%d matches)", i, matches)
	})
	for fa, err := range fasta.File(*qFile) {
		if err != nil {
			return err
		}
		s := sketching.Sketch(fa.Sequence, kmerLen, sk.scale)
		for _, f := range idx.Search(s) {
			var sim float64
			if useMyDist {
				sim = 1 - myDist(s, sk.skch[f], len(fa.Sequence), sk.lens[f])
			} else {
				sim = 1 - mash.FromJaccard(jaccard(s, sk.skch[f]), kmerLen)
			}
			if sim >= *minSim {
				matches++
				output := []string{
					fmt.Sprintf("%.0f%%", sim*100),
					string(fa.Name),
					sk.names[f],
				}
				out.Write(output)
			}
		}
		pt.Inc()
	}
	pt.Done()

	return nil
}
