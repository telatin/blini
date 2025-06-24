// Sketching logic.

package main

import (
	"fmt"
	"strings"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
)

// Main function for sketching operation.
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
