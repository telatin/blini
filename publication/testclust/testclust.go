// Tests clustering results of blini and mmseqs.
package main

import (
	"encoding/csv"
	"fmt"
	"regexp"

	"github.com/fluhus/gostuff/clustering"
	"github.com/fluhus/gostuff/iterx"
	"github.com/fluhus/gostuff/jio"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/maps"
)

func main() {
	fmt.Println("Testing mmseqs clusters")
	groups, err := readMMSClusters("tmp.mm_cluster.tsv")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(groups), "clusters")

	gotTags := toTags(groups)
	wantTags := toTags(toWantGroups(groups))
	fmt.Println("ARI:", clustering.AdjustedRandIndex(gotTags, wantTags))

	fmt.Println("Testing blini clusters")
	for _, scale := range []string{"200", "100", "50", "25"} {
		bln := struct {
			ByName [][]string
		}{}
		if err := jio.Read("tmp.blini_"+scale+".json", &bln); err != nil {
			panic(err)
		}
		groups = bln.ByName

		gotTags = toTags(groups)
		wantTags = toTags(toWantGroups(groups))
		fmt.Printf("[%s] ARI: %f\n", scale,
			clustering.AdjustedRandIndex(gotTags, wantTags))
	}
}

// Returns the clusters from MMseq's TSV output.
func readMMSClusters(file string) ([][]string, error) {
	m := map[string][]string{}
	for line, err := range iterx.CSVFile(file, toTSV) {
		if err != nil {
			return nil, err
		}
		m[line[0]] = append(m[line[0]], line[1])
	}
	return maps.Values(m), nil
}

// Makes a CSV reader a TSV reader.
func toTSV(r *csv.Reader) {
	r.Comma = '\t'
}

// Returns the cluster index number for each element.
func toTags(groups [][]string) []int {
	m := map[string]int{}
	for i, g := range groups {
		for _, s := range g {
			m[s] = i
		}
	}
	return snm.SliceToSlice(snm.Sorted(maps.Keys(m)), func(k string) int {
		return m[k]
	})
}

// Returns the "real" grouping.
func toWantGroups(groups [][]string) [][]string {
	re := regexp.MustCompile(`^.*\.`)
	m := map[string][]string{}
	for _, g := range groups {
		for _, s := range g {
			k := re.FindString(s)
			if k == "" {
				panic(fmt.Sprintf("could not match element: %q", s))
			}
			m[k] = append(m[k], s)
		}
	}
	return maps.Values(m)
}
