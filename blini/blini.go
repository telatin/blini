package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fluhus/biostuff/mash/v2"
	"github.com/fluhus/blini/sketching"
)

/*
TODO
- Make ptimer output to stdout
- Tests for sketching and for common
- Grouping by file, by regex?
*/

const (
	kmerLen  = 21
	idxScale = 4

	useMyDist    = true          // Use a new experiemental distance func.
	indexSuffix  = ".blini"      // Suffix of pre-sketched files.
	unmatchedRef = "(unmatched)" // The "reference" value of an unmatched query.
)

var (
	qFile     = flag.String("q", "", "Query file")
	rFile     = flag.String("r", "", "Reference file")
	oFile     = flag.String("o", "", "Output file or prefix")
	contn     = flag.Bool("c", false, "Use containment rather than full match")
	minSim    = flag.Float64("m", 0.9, "Minimum similarity for match")
	scale     = flag.Uint64("s", 100, "Use 1/`scale` of the kmers")
	unmatched = flag.Bool("u", false, "Include unmatched queries in search output")
	showVer   = flag.Bool("version", false, "Print version and exit")
	version   = "0.4.1"
)

func main() {
	flag.Parse()
	if *showVer {
		fmt.Printf("Blini %s\n", version)
		return
	}
	debug.SetGCPercent(20)

	var err error
	if *qFile != "" && *rFile != "" {
		err = mainSearch()
	} else if *qFile != "" {
		err = mainCluster()
	} else if *rFile != "" {
		err = mainSketch()
	} else {
		fmt.Printf("Blini (%s)\n\n", version)
		fmt.Println("Please select -q for clustering, -r for sketching,",
			"or both for searching.")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}
}

// Returns the jaccard/containment jaccard.
func jaccard(a, b []uint64) float64 {
	if *contn {
		return sketching.Containment(a, b)
	} else {
		return sketching.Jaccard(a, b)
	}
}

// Returns a specialized distance.
func myDist(a, b []uint64, alen, blen int) float64 {
	if *contn {
		return mash.FromJaccard(sketching.Containment(a, b), kmerLen)
	} else {
		return sketching.MyDist(a, b, alen, blen, kmerLen)
	}
}
