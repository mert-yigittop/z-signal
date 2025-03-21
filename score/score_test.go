package score

import (
	"github.com/subspace-343/z-score/order"
	"math/rand"
	"testing"
	"time"
)

func generateRandomLevels(n int) []order.Level {
	rand.Seed(time.Now().UnixNano())
	levels := make([]order.Level, n)
	for i := range levels {
		levels[i] = order.Level{
			Price:    rand.Float64() * 10000,
			Quantity: rand.Float64() * 10,
		}
	}
	return levels
}

func generateOrderBook(n int) *order.Book {
	return &order.Book{
		Bids: generateRandomLevels(n),
		Asks: generateRandomLevels(n),
	}
}

func extractQuantities(levels []order.Level) []float64 {
	q := make([]float64, len(levels))
	for i, l := range levels {
		q[i] = l.Quantity
	}
	return q
}

func BenchmarkCalculateStdDevAndMean(b *testing.B) {
	levels := generateRandomLevels(1000)
	s := &Score{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.calculateStdDevAndMean(levels)
	}
}

func BenchmarkCalculateSizeZScore(b *testing.B) {
	orderBook := generateOrderBook(1000)
	s := &Score{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.calculateSizeZScore(orderBook)
	}
}
