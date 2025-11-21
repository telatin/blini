package simul

import (
	"bytes"
	"slices"
	"testing"

	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/maps"
)

func TestMutNuc(t *testing.T) {
	for _, b := range []byte("ATCG") {
		m := map[byte]int{}
		for range 1000 {
			m[MutNuc(b)]++
		}
		if len(m) != 3 {
			t.Errorf("MutNuc(%q) got %d chars, want 3", b, len(m))
		}
		for _, c := range []byte("ATCG") {
			if c == b {
				if m[c] > 0 {
					t.Errorf("MutNuc(%q)=%q", c, c)
				}
				continue
			}
			if m[c] < 280 {
				t.Errorf("MutNuc(%q)=%q count=%d, want 280",
					b, c, m[c])
			}
		}
	}
}

func TestRandSeq(t *testing.T) {
	s := RandSeq(1000)
	if len(s) != 1000 {
		t.Fatalf("RandSeq(1000) length is %d, want 1000",
			len(s))
	}
	m := map[byte]int{}
	for _, b := range s {
		m[b]++
	}
	keys := string(snm.Sorted(maps.Keys(m)))
	if keys != "ACGT" {
		t.Fatalf("RandSeq(1000) letters are %q, want %q",
			keys, "ACGT")
	}
	for k, v := range m {
		if v < 200 {
			t.Errorf("RandSeq(1000) count[%q]=%d, want 200",
				k, v)
		}
	}
}

func TestMutSeqPerc(t *testing.T) {
	s := RandSeq(1000)
	for _, n := range []int{1, 10, 30, 50} {
		c := slices.Clone(s)
		mut := MutSeqPerc(c, n)
		if !bytes.Equal(s, c) {
			t.Fatalf("MutSeq(...) modified source")
		}
		got, want := 0, n*10
		for i := range mut {
			if mut[i] != c[i] {
				got++
			}
		}
		if got != want {
			t.Errorf("MutSeq(\"...\",%d) mutated %d, want %d",
				n, got, want)
		}
	}
}
