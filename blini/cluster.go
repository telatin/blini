// Clustering logic.

package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/biostuff/mash/v2"
	"github.com/fluhus/blini/sketching"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/jio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
)

// Main function for clustering operation.
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
	perm := sortedPerm(sk.lens, func(a, b int) int {
		return cmp.Compare(b, a)
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
			sim := 1 - mash.FromJaccard(jaccard(sk.skch[f], s), kmerLen)
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

	{ // TODO(fluhus): Organize this code.
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
		slices.Sort(c[1:]) // First element is the representative.
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
		fmt.Println("Generating output")
		// JSON output.
		output := map[string]any{
			"byNumber": clusters,
			"byName":   byName,
		}
		if err := jio.Write(*oFile+".json", output); err != nil {
			return err
		}

		// Fasta output.
		reps := snm.SliceToSlice(clusters, func(a []int) int { return a[0] })
		fout, err := aio.Create(*oFile + ".fasta")
		if err != nil {
			return err
		}
		defer fout.Close()
		i := -1
		for fa, err := range fasta.File(*qFile) {
			if err != nil {
				return err
			}
			if len(reps) == 0 {
				break
			}
			i++
			if i == reps[0] {
				if err := fa.Write(fout); err != nil {
					return err
				}
				reps = reps[1:]
			}
		}
	} else {
		fmt.Println("No output")
	}

	return nil
}

// Returns the indexes of slice elements if they were sorted.
func sortedPerm[T any](s []T, cmp func(T, T) int) []int {
	return snm.SortedFunc(
		snm.Slice(len(s), func(i int) int { return i }),
		func(i, j int) int { return cmp(s[i], s[j]) },
	)
}
