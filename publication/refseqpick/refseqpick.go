// Picks out random genomes from an input reference.
package main

import (
	"fmt"
	"os"

	"github.com/fluhus/biostuff/formats/fasta"
	"github.com/fluhus/blini/publication/simul"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/reservoir"
)

func main() {
	inFile := os.Args[1]

	r := reservoir.New[*fasta.Fasta](100)
	for fa, err := range fasta.File(inFile) {
		if err != nil {
			panic(err)
		}
		r.Add(fa)
	}

	f, err := aio.Create("publication/testdata/vir_all.fa")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fm, err := aio.Create("publication/testdata/mut_all.fa")
	if err != nil {
		panic(err)
	}
	defer fm.Close()

	for i, fa := range r.Elements {
		txt, _ := fa.MarshalText()
		file := fmt.Sprintf("publication/testdata/vir_%d.fa", i+1)
		os.WriteFile(file, txt, 0o600)
		f.Write(txt)

		fa.Sequence = simul.MutSeqPerc(fa.Sequence, 1)
		txt, _ = fa.MarshalText()
		file = fmt.Sprintf("publication/testdata/mut_%d.fa", i+1)
		os.WriteFile(file, txt, 0o600)
		fm.Write(txt)
	}
}
