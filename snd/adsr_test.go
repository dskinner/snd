package snd

import (
	"testing"
	"time"
)

func BenchmarkADSR(b *testing.B) {
	adsr := NewADSR(100*time.Second, 200*time.Second, 300*time.Second, 400*time.Second, 0.7, 1, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		adsr.Prepare(uint64(n))
	}
}
