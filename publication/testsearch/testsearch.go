// Tests search results of blini and sourmash.
package main

import (
	"fmt"
	"path/filepath"

	"github.com/fluhus/gostuff/csvdec"
)

func main() {
	e, err := readBlini("tmp.blini_vir.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println(findMatches(e))

	e, err = readBlini("tmp.blini_mut.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println(findMatches(e))

	e, err = readSourmash("tmp.sm_vir_*.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println(findMatches(e))

	e, err = readSourmash("tmp.sm_mut_*.csv")
	if err != nil {
		panic(err)
	}
	fmt.Println(findMatches(e))
}

// Reads matches produced by Blini.
func readBlini(file string) ([][2]string, error) {
	type entry struct {
		Query, Reference string
	}
	var result [][2]string
	for e, err := range csvdec.FileHeader[entry](file, nil) {
		if err != nil {
			return nil, err
		}
		result = append(result, [2]string{e.Query, e.Reference})
	}
	return result, nil
}

// Reads matches produced by Sourmash.
func readSourmash(glob string) ([][2]string, error) {
	type entry struct {
		Query     string `csvdec:"query_name"`
		Reference string `csvdec:"name"`
	}
	var result [][2]string
	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		for e, err := range csvdec.FileHeader[entry](file, nil) {
			if err != nil {
				return nil, err
			}
			result = append(result, [2]string{e.Query, e.Reference})
		}
	}
	return result, nil
}

// Returns the number of matches (query=reference)
// and the number of mismatches.
func findMatches(entries [][2]string) (int, int) {
	matches, mismatches := 0, 0
	for _, e := range entries {
		if e[0] == e[1] {
			matches++
		} else {
			mismatches++
		}
	}
	return matches, mismatches
}
