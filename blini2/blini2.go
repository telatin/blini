package main

import (
	"cmp"
	"flag"
	"fmt"
	"io"
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
- Flag for scale?
- Break down this file for each main
- Tests for sketching and for common
- Grouping by file, by regex?
*/

const (
	kmerLen = 21
	scale   = 50

	useMyDist   = true     // Use a new experiemental distance func.
	indexSuffix = ".blini" // Suffix of pre-sketched files.
)

var (
	qFile  = flag.String("q", "", "Query file")
	rFile  = flag.String("r", "", "Reference file")
	oFile  = flag.String("o", "", "Output file or prefix")
	contn  = flag.Bool("c", false, "Use containment rather than full match")
	minSim = flag.Float64("s", 0.9, "Minimum similarity for match")
)

func main() {
	flag.Parse()
	fmt.Println("Let's go!")
	fmt.Println("K:      ", kmerLen)
	fmt.Println("Min sim:", *minSim)
	fmt.Println("Scale:  ", scale)
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
	var skch [][]uint64
	var lens []int
	var names []string
	var err error
	if strings.HasSuffix(*rFile, indexSuffix) {
		fmt.Println("Reading prepared sketches")
		skch, lens, names, err = readSketches(*rFile)
	} else {
		fmt.Println("Sketching reference sequences")
		skch, lens, names, err = sketchFile(*rFile)
	}
	if err != nil {
		return err
	}

	fmt.Println("Indexing")
	pt := ptimer.New()
	idx := sketching.NewIndex(scale * 10)
	for i, s := range skch {
		idx.Add(s, i)
		pt.Inc()
	}
	pt.Done()

	fmt.Println("Searching")
	var matches []string
	pt = ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d (%d matched)", i, len(matches))
	})
	for fa, err := range fasta.File(*qFile) {
		if err != nil {
			return err
		}
		s := sketching.Sketch(fa.Sequence, kmerLen, scale)
		for _, f := range idx.Search(s) {
			sim := 1 - mash.FromJaccard(dist(s, skch[f]), kmerLen)
			if useMyDist {
				sim = 1 - myDist(s, skch[f], len(fa.Sequence), lens[f])
			}
			if sim >= *minSim {
				matches = append(matches, fmt.Sprintf(
					"(%.0f%%) %s >>>>> %s",
					sim*100, fa.Name, names[f]))
			}
		}
		pt.Inc()
	}
	pt.Done()

	for _, m := range matches {
		fmt.Println(m)
	}

	return nil
}

func mainCluster() error {
	fmt.Println("--------------------")
	fmt.Println("CLUSTERING OPERATION")
	fmt.Println("--------------------")
	fmt.Println("Sketching sequences")
	skch, lens, names, err := sketchFile(*qFile)
	if err != nil {
		return err
	}

	fmt.Println("Indexing")
	pt := ptimer.New()
	idx := sketching.NewIndex(scale * 10)
	for i, s := range skch {
		idx.Add(s, i)
		pt.Inc()
	}
	pt.Done()

	idx.Clean()

	fmt.Println("Clustering")
	perm := sortedPerm(len(lens), func(i, j int) int {
		return cmp.Compare(lens[j], lens[i])
	})
	friends := 0
	var clusters [][]int
	pt = ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d (%dc %df)", i, len(clusters), friends/i)
	})
	for _, i := range perm {
		s := skch[i]
		if s == nil {
			pt.Inc()
			continue
		}
		skch[i] = nil
		fr := idx.Search(s)
		friends += len(fr)

		// Create cluster.
		c := []int{i}
		for _, f := range fr {
			if skch[f] == nil {
				continue
			}
			sim := 1 - mash.FromJaccard(dist(skch[f], s), kmerLen)
			if useMyDist {
				sim = 1 - myDist(skch[f], s, lens[f], lens[i])
			}
			if sim < *minSim {
				continue
			}
			c = append(c, f)
			skch[f] = nil
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
			return names[i]
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
	fmt.Println("Sketching sequences")
	skch, lens, names, err := sketchFile(*rFile)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(*oFile, indexSuffix) {
		*oFile += indexSuffix
	}
	fmt.Println("Saving to:", *oFile)
	f, err := aio.Create(*oFile)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := range skch {
		err := bnry.Write(f, skch[i], lens[i], names[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func sketchFile(file string) ([][]uint64, []int, []string, error) {
	pt := ptimer.New()
	var skch [][]uint64
	var lens []int
	var names []string
	for fa, err := range fasta.File(file) {
		if err != nil {
			return nil, nil, nil, err
		}
		skch = append(skch, sketching.Sketch(fa.Sequence, kmerLen, scale))
		lens = append(lens, len(fa.Sequence))
		names = append(names, string(fa.Name))
		pt.Inc()
	}
	pt.Done()

	return skch, lens, names, nil
}

func readSketches(file string) ([][]uint64, []int, []string, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()
	var skch [][]uint64
	var lens []int
	var names []string

	for {
		var s []uint64
		var l int
		var n string
		err := bnry.Read(f, &s, &l, &n)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, nil, err
		}
		skch = append(skch, s)
		lens = append(lens, l)
		names = append(names, n)
	}

	return skch, lens, names, err
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
