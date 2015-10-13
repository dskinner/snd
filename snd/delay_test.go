package snd

import (
	"testing"
	"time"
)

func BenchmarkDelay(b *testing.B) {
	dly := NewDelay(100*time.Millisecond, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dly.Prepare(uint64(n))
	}
}

func BenchmarkDelayComb(b *testing.B) {
	cmb := NewComb(0.8, 100*time.Millisecond, newunit())
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cmb.Prepare(uint64(n))
	}
}
