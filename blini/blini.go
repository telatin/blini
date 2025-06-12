package main

import (
	"bytes"
	"cmp"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"iter"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/biostuff/mash/v2"
	"github.com/fluhus/blini/sketching"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/jio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/maps"
)

/*
TODO
- Document functions
- Add representative output to cluster
- Finalize output of search
- Break down this file for each main
- Tests for sketching and for common
- Grouping by file, by regex?
*/

const (
	kmerLen  = 21
	idxScale = 5

	useMyDist   = true     // Use a new experiemental distance func.
	indexSuffix = ".blini" // Suffix of pre-sketched files.
)

var (
	qFile  = flag.String("q", "", "Query file")
	rFile  = flag.String("r", "", "Reference file")
	oFile  = flag.String("o", "", "Output file or prefix")
	contn  = flag.Bool("c", false, "Use containment rather than full match")
	minSim = flag.Float64("m", 0.9, "Minimum similarity for match")
	scale  = flag.Uint64("s", 200, "Use 1/`ratio` of the kmers")
)

func main() {
	flag.Parse()
	debug.SetGCPercent(20)

	var err error
	if *qFile != "" && *rFile != "" {
		err = mainSearch()
	} else if *qFile != "" {
		err = mainCluster()
	} else if *rFile != "" {
		err = mainSketch()
	}
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}
}

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
	buf := bytes.NewBuffer(nil)
	out := csv.NewWriter(buf)
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
			sim := 1 - mash.FromJaccard(dist(s, sk.skch[f]), kmerLen)
			if useMyDist {
				sim = 1 - myDist(s, sk.skch[f], len(fa.Sequence), sk.lens[f])
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

	out.Flush()
	fmt.Printf("%s", buf.Bytes())

	return nil
}

func mainCluster() error {
	fmt.Println("--------------------")
	fmt.Println("CLUSTERING OPERATION")
	fmt.Println("--------------------")
	fmt.Println("Scale:", *scale)
	fmt.Println("Min sim:", *minSim)

	fmt.Println("Sketching sequences")
	sk, err := collectSketches(sketchFile(*qFile))
	if err != nil {
		return err
	}

	fmt.Println("Indexing")
	pt := ptimer.New()
	idx := sketching.NewIndex(*scale * idxScale)
	for i, s := range sk.skch {
		idx.Add(s, i)
		pt.Inc()
	}
	pt.Done()

	idx.Clean()

	fmt.Println("Clustering")
	perm := sortedPerm(len(sk.lens), func(i, j int) int {
		return cmp.Compare(sk.lens[j], sk.lens[i])
	})
	friends := 0
	var clusters [][]int
	pt = ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d (%dc %df)", i, len(clusters), friends/i)
	})
	for _, i := range perm {
		s := sk.skch[i]
		if s == nil {
			pt.Inc()
			continue
		}
		sk.skch[i] = nil
		fr := idx.Search(s)
		friends += len(fr)

		// Create cluster.
		c := []int{i}
		for _, f := range fr {
			if sk.skch[f] == nil {
				continue
			}
			sim := 1 - mash.FromJaccard(dist(sk.skch[f], s), kmerLen)
			if useMyDist {
				sim = 1 - myDist(sk.skch[f], s, sk.lens[f], sk.lens[i])
			}
			if sim < *minSim {
				continue
			}
			c = append(c, f)
			sk.skch[f] = nil
		}
		clusters = append(clusters, c)
		pt.Inc()
	}
	pt.Done()

	{
		alli := sets.Set[int]{}
		lens := 0
		for _, c := range clusters {
			alli.Add(c...)
			lens += len(c)
		}
		if lens != len(alli) {
			fmt.Println("lens!=alli:", lens, len(alli))
		}
		if !maps.Equal(alli, sets.Of(perm...)) {
			fmt.Println("bad alli")
		}
	}

	// Sort clusters for deterministic output.
	for _, c := range clusters {
		slices.Sort(c)
	}
	slices.SortFunc(clusters, func(a, b []int) int {
		return cmp.Compare(a[0], b[0])
	})

	// Create clusters by names.
	byName := snm.SliceToSlice(clusters, func(c []int) []string {
		return snm.SliceToSlice(c, func(i int) string {
			return sk.names[i]
		})
	})

	if *oFile != "" {
		if err := jio.Write(*oFile+"_bynumber.json", clusters); err != nil {
			return err
		}
		if err := jio.Write(*oFile+"_byname.json", byName); err != nil {
			return err
		}
	} else {
		fmt.Println("No output")
	}

	return nil
}

func mainSketch() error {
	fmt.Println("----------------")
	fmt.Println("SKETCH OPERATION")
	fmt.Println("----------------")
	fmt.Println("Scale:", *scale)

	if !strings.HasSuffix(*oFile, indexSuffix) {
		*oFile += indexSuffix
	}
	fmt.Println("Saving to:", *oFile)
	f, err := aio.Create(*oFile)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Println("Sketching sequences")
	pt := ptimer.New()
	for e, err := range sketchFile(*rFile) {
		if err != nil {
			return err
		}
		if err := bnry.Write(f, e.s, e.ln, e.name, e.scale); err != nil {
			return err
		}
		pt.Inc()
	}
	pt.Done()
	return nil
}

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

func dist(a, b []uint64) float64 {
	if *contn {
		return sketching.Containment(a, b)
	} else {
		return sketching.Jaccard(a, b)
	}
}

func myDist(a, b []uint64, alen, blen int) float64 {
	if *contn {
		return mash.FromJaccard(sketching.Containment(a, b), kmerLen)
	} else {
		return sketching.MyDist(a, b, alen, blen, kmerLen)
	}
}

func sortedPerm(n int, cmp func(i, j int) int) []int {
	return snm.SortedFunc(snm.Slice(n, func(i int) int { return i }), cmp)
}

type sketches struct {
	skch  [][]uint64
	lens  []int
	names []string
	scale uint64
}

type sketchEntry struct {
	s     []uint64
	ln    int
	name  string
	scale uint64
}

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
