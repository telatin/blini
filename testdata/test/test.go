// Tests the outputs of running Blini on the test data.
package main

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/gostuff/csvx"
	"github.com/fluhus/gostuff/jio"
)

const (
	dataDir    = "testdata"
	resultsDir = "tmp"
)

func main() {
	if err := testClusters(); err != nil {
		panic(err)
	}
	if err := testClustersSeqs(); err != nil {
		panic(err)
	}
	if err := testSearch(); err != nil {
		panic(err)
	}
	fmt.Println("OK!")
}

// Tests that all sequences went to the right cluster.
func testClusters() error {
	var result struct {
		ByName [][]string
	}
	if err := jio.Read(resultsDir+"/clust.json", &result); err != nil {
		return err
	}
	got := result.ByName
	sortClusters(got)
	want := [][]string{
		{"ref1", "ref1.1", "ref1.2", "ref1.3"},
		{"ref2", "ref2.1", "ref2.2", "ref2.3"},
		{"ref3", "ref3.1", "ref3.2", "ref3.3"},
	}
	if !slices.EqualFunc(got, want, slices.Equal) {
		return fmt.Errorf("wrong clustering result: got %v, want %v",
			got, want)
	}
	return nil
}

// Tests that the output sequences are the same as input.
func testClustersSeqs() error {
	// Read original sequences.
	want := map[string]string{}
	for fa, err := range fasta.File(dataDir + "/clust.fa.zst") {
		if err != nil {
			return err
		}
		want[string(fa.Name)] = string(fa.Sequence)
	}

	// Test result sequences.
	namePattern := regexp.MustCompile(`^ref\d$`)
	for fa, err := range fasta.File(resultsDir + "/clust.fasta") {
		if err != nil {
			return err
		}
		if !namePattern.Match(fa.Name) {
			return fmt.Errorf("unexpected name: %q", fa.Name)
		}
		if string(fa.Sequence) != want[string(fa.Name)] {
			return fmt.Errorf("mismatching sequence for %q", fa.Name)
		}
	}

	return nil
}

// Tests that each query was matched with its reference.
func testSearch() error {
	const wantSimilarity = 98

	type entry struct {
		Similarity, Query, Reference string
	}

	qPattern := regexp.MustCompile(`^query(\d)\.\d$`)
	rPattern := regexp.MustCompile(`^ref(\d)$`)
	similarities := map[string]bool{ // Accepted similarity values.
		fmt.Sprint(wantSimilarity, "%"):   true,
		fmt.Sprint(wantSimilarity+1, "%"): true,
		fmt.Sprint(wantSimilarity-1, "%"): true,
	}

	for row, err := range csvx.DecodeFileHeader[entry](resultsDir + "/search.csv") {
		if err != nil {
			return err
		}
		q := qPattern.FindStringSubmatch(row.Query)
		r := rPattern.FindStringSubmatch(row.Reference)
		if q == nil {
			return fmt.Errorf("unexpected query name: %q", row.Query)
		}
		if r == nil {
			return fmt.Errorf("unexpected reference name: %q", row.Reference)
		}
		if q[1] != r[1] {
			return fmt.Errorf("mismatching query and reference: %q %q",
				row.Query, row.Reference)
		}
		if !similarities[row.Similarity] {
			return fmt.Errorf("unexpected similarity: %s", row.Similarity)
		}
	}
	return nil
}

// Sorts a clustering assignment to make clusterings comparable.
func sortClusters[T cmp.Ordered](a [][]T) {
	for _, x := range a {
		// Skip first element to keep cluster representative first.
		slices.Sort(x[1:])
	}
	slices.SortFunc(a, func(x, y []T) int {
		return cmp.Compare(x[0], y[0])
	})
}
