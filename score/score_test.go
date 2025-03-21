package score

import (
	"math/rand"
	"testing"
	"time"
)

func generateRandomLevels(n int) []Level {
	rand.Seed(time.Now().UnixNano())
	levels := make([]Level, n)
	for i := range levels {
		levels[i] = Level{
			Price:    rand.Float64() * 10000,
			Quantity: rand.Float64() * 10,
		}
	}
	return levels
}

func generateOrderBook(n int) *OrderBook {
	return &OrderBook{
		Bids: generateRandomLevels(n),
		Asks: generateRandomLevels(n),
	}
}

func extractQuantities(levels []Level) []float64 {
	q := make([]float64, len(levels))
	for i, l := range levels {
		q[i] = l.Quantity
	}
	return q
}

func BenchmarkCalculateStdDevAndMean(b *testing.B) {
	levels := generateRandomLevels(1000)
	s := &Score{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = s.calculateStdDevAndMean(levels)
	}
}

func BenchmarkCalculateSizeZScore(b *testing.B) {
	orderBook := generateOrderBook(1000)
	s := &Score{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.calculateSizeZScore(orderBook)
	}
}
