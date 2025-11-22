package sketching

import "testing"

func TestJaccard(t *testing.T) {
	tests := []struct {
		a, b []uint64
		want float64
	}{
		{[]uint64{1}, nil, 0},
		{[]uint64{1}, []uint64{1}, 1.0 / (1 + ghostUnion)},
		{[]uint64{1}, []uint64{1, 2}, 1.0 / (2 + ghostUnion)},
		{[]uint64{3}, []uint64{1, 2}, 0},
		{[]uint64{1, 2, 3}, []uint64{6, 7, 8}, 0},
		{[]uint64{1, 2, 4}, []uint64{0, 2, 3, 4, 5}, 2.0 / (6 + ghostUnion)},
	}

	for _, test := range tests {
		if got := Jaccard(test.a, test.b); got != test.want {
			t.Errorf("Jaccard(%v,%v)=%v, want %v",
				test.a, test.b, got, test.want)
		}
		if got := Jaccard(test.b, test.a); got != test.want {
			t.Errorf("Jaccard(%v,%v)=%v, want %v",
				test.b, test.a, got, test.want)
		}
	}
}
