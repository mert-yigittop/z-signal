package score

import (
	"context"
	"fmt"
	"github.com/starbase-343/ferengi/utils/multiplexer"
	"github.com/starbase-343/ferengi/utils/streamer"
	"github.com/subspace-343/z-score/order"
	"github.com/subspace-343/z-score/printer"
	"log"
	"math"
	"sync/atomic"
)

var (
	ErrLevelCountMismatch          = fmt.Errorf("level count mismatch")
	ErrStdDevMustBeGreaterThanZero = fmt.Errorf("standard deviation must be greater than 0")
)

var (
	// LevelCount is the number of levels to calculate the z-score
	LevelCount = 20
)

type Score struct {
	running     atomic.Bool
	healthCheck atomic.Bool
	printer     printer.Printer
}

func NewScore(p printer.Printer) *Score {
	return &Score{
		printer: p,
	}
}

func (s *Score) AsyncRun(ctx context.Context, multiplexer multiplexer.Multiplexer, streamerCount int) {
	if s.running.Load() {
		log.Println("warning: mm runner is already running")
		return
	}
	s.running.Store(true)
	defer s.running.Store(false)

	ticksCh, err := multiplexer.Subscribe("z-score")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ticks := <-ticksCh:
			if len(ticks) == streamerCount {
				s.processTicks(ticks, streamerCount)

				if !s.healthCheck.Load() {
					log.Println("[mm runner] receiving ticks and operational")
					s.healthCheck.Store(true)
				}
			}
		}
	}
}

func (s *Score) processTicks(ticks []streamer.Tick, streamerCount int) {
	for _, tick := range ticks {
		if err := s.calculate(tick); err != nil {
			log.Println(err)
		}
	}
}

func (s *Score) calculate(tick streamer.Tick) error {
	asks := make([]order.Level, len(tick.AskLevels))
	bids := make([]order.Level, len(tick.BidLevels))

	if len(asks) != LevelCount || len(bids) != LevelCount {
		return ErrLevelCountMismatch
	}

	for i, ask := range tick.AskLevels {
		asks[i] = order.Level{Price: ask.Price, Quantity: ask.Quantity}
	}
	for i, bid := range tick.BidLevels {
		bids[i] = order.Level{Price: bid.Price, Quantity: bid.Quantity}
	}

	//sort.Slice(asks, func(i, j int) bool { return asks[i].Price < asks[j].Price })
	//sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

	ob := order.Book{Asks: asks, Bids: bids}
	if err := s.calculateSizeZScore(&ob); err != nil {
		return err
	}

	return nil
}

func (s *Score) calculateStdDevAndMean(l []order.Level) (float64, float64) {
	var sum, sqSum float64
	for i := 0; i < LevelCount; i++ {
		q := l[i].Quantity
		sum += q
		sqSum += q * q
	}

	mean := sum / float64(LevelCount)
	variance := (sqSum / float64(LevelCount)) - (mean * mean)
	if variance == 0 {
		return 0, mean
	}

	return math.Sqrt(variance), mean
}

func (s *Score) calculateSizeZScore(ob *order.Book) error {
	bidsStdDev, bidsMean := s.calculateStdDevAndMean(ob.Bids)
	asksStdDev, asksMean := s.calculateStdDevAndMean(ob.Asks)

	if bidsStdDev == 0 || asksStdDev == 0 {
		return ErrStdDevMustBeGreaterThanZero
	}

	for i := 0; i < LevelCount; i++ {
		ob.Bids[i].ZScore = (ob.Bids[i].Quantity - bidsMean) / bidsStdDev
		ob.Asks[i].ZScore = (ob.Asks[i].Quantity - asksMean) / asksStdDev
	}

	//s.printer.Print(*ob)
	return nil
}
