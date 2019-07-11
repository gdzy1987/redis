package topology

import "testing"

func BenchmarkNewUUid(t *testing.B) {
	for i := 0; i < t.N; i++ {
		x := NewSUID()
		_ = x
	}
}
