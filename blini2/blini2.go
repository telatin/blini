package main

import (
	"cmp"
	"flag"
	"fmt"
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

const (
	kmerLen = 21
	minSim  = 0.9
	scale   = 20

	compensatingContainment = true     // Punish containment for possible mismatches.
	quickCommon             = true     // Use a quick version of "common"
	indexSuffix             = ".blini" // Suffix of pre-sketched files.
)

var (
	qFile = flag.String("q", "", "Query file")
	rFile = flag.String("r", "", "Reference file")
	oFile = flag.String("o", "", "Output file or prefix")
	contn = flag.Bool("c", false, "Use containment rather than full match")
)

func main() {
	flag.Parse()
	fmt.Println("Let's go!")
	fmt.Println("K:      ", kmerLen)
	fmt.Println("Min sim:", minSim)
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
	var names []string
	var err error
	if strings.HasSuffix(*rFile, indexSuffix) {
		fmt.Println("Reading prepared sketches")
		skch, names, err = readSketches(*rFile)
	} else {
		fmt.Println("Sketching reference sequences")
		skch, names, err = sketchFile(*rFile)
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
			if sim >= minSim {
				matches = append(matches, fmt.Sprintf(
					"(%.0f%%) %s >>>>> %s",
					sim*100, fa.Name, names[f]))
				// break
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
	skch, names, err := sketchFile(*qFile)
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
	perm := orderBySize(skch)
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
			if sim < minSim {
				continue
			}
			c = append(c, f)
			skch[f] = nil
		}
		clusters = append(clusters, c)
		pt.Inc()
	}
	pt.Done()

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
	skch, names, err := sketchFile(*rFile)
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
	if err := bnry.Write(f, len(skch)); err != nil {
		return err
	}
	for _, s := range skch {
		if err := bnry.Write(f, s); err != nil {
			return err
		}
	}
	if err := bnry.Write(f, names); err != nil {
		return err
	}
	return nil
}

func sketchFile(file string) ([][]uint64, []string, error) {
	pt := ptimer.New()
	var skch [][]uint64
	var names []string
	for fa, err := range fasta.File(file) {
		if err != nil {
			return nil, nil, err
		}
		skch = append(skch, sketching.Sketch(fa.Sequence, kmerLen, scale))
		names = append(names, string(fa.Name))
		pt.Inc()
	}
	pt.Done()

	return skch, names, nil
}

func readSketches(file string) ([][]uint64, []string, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	var n int
	if err := bnry.Read(f, &n); err != nil {
		return nil, nil, err
	}
	skch := make([][]uint64, n)
	for i := range skch {
		var s []uint64
		if err := bnry.Read(f, &s); err != nil {
			return nil, nil, err
		}
		skch[i] = s
	}
	var names []string
	if err := bnry.Read(f, &names); err != nil {
		return nil, nil, err
	}
	return skch, names, err
}

func common(a, b []uint64) int {
	if quickCommon {
		c, i, j := 0, 0, 0
		for i < len(a) && j < len(b) {
			if a[i] < b[j] {
				i++
			} else if b[j] < a[i] {
				j++
			} else {
				c++
				i++
				j++
			}
		}
		return c
	}
	return len(sets.Of(a...).Intersect(sets.Of(b...)))
}

func jaccard(a, b []uint64) float64 {
	i := common(a, b)
	u := len(a) + len(b) - i
	return float64(i) / float64(u)
}

func containment(a, b []uint64) float64 {
	i := common(a, b)
	if compensatingContainment {
		return float64(i) / float64(len(a)+(len(a)-i))
	} else {
		return float64(i) / float64(len(a))
	}
}

func dist(a, b []uint64) float64 {
	if *contn {
		return containment(a, b)
	} else {
		return jaccard(a, b)
	}
}

func orderBySize(a [][]uint64) []int {
	perm := snm.Slice(len(a), func(i int) int { return i })
	slices.SortFunc(perm, func(i, j int) int {
		return cmp.Compare(len(a[j]), len(a[i]))
	})
	return perm
}
