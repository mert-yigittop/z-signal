package score

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/starbase-343/ferengi/utils/multiplexer"
	"github.com/starbase-343/ferengi/utils/streamer"
	"log"
	"math"
	"sync/atomic"
)

var (
	ErrLevelCountMustBeGreaterThanZero = fmt.Errorf("level count must be greater than 0")
	ErrStdDevMustBeGreaterThanZero     = fmt.Errorf("standard deviation must be greater than 0")
)

var (
	// LevelCount is the number of levels to calculate the z-score
	LevelCount = 20

	// OutliersDetection is the threshold for the z-score to detect outliers
	OutliersDetection = 1.5
)

type Score struct {
	running     atomic.Bool
	healthCheck atomic.Bool
}

func NewScore() *Score {
	return &Score{}
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

type Level struct {
	Price, Quantity, ZScore float64
}
type OrderBook struct {
	Asks []Level
	Bids []Level
}

func (s *Score) processTicks(ticks []streamer.Tick, streamerCount int) {
	for _, tick := range ticks {
		if err := s.calculate(tick); err != nil {
			log.Println(err)
		}
	}
}

func (s *Score) calculate(tick streamer.Tick) error {
	asks := make([]Level, len(tick.AskLevels))
	bids := make([]Level, len(tick.BidLevels))

	for i, ask := range tick.AskLevels {
		asks[i] = Level{Price: ask.Price, Quantity: ask.Quantity}
	}
	for i, bid := range tick.BidLevels {
		bids[i] = Level{Price: bid.Price, Quantity: bid.Quantity}
	}

	//sort.Slice(asks, func(i, j int) bool { return asks[i].Price < asks[j].Price })
	//sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

	ob := OrderBook{Asks: asks, Bids: bids}
	if err := s.calculateSizeZScore(&ob); err != nil {
		return err
	}

	return nil
}

func (s *Score) calculateStdDevAndMean(l []Level) (float64, float64) {
	count := min(LevelCount, len(l))
	if count == 0 {
		return 0, 0
	}

	var sum, sqSum float64
	for i := 0; i < count; i++ {
		q := l[i].Quantity
		sum += q
		sqSum += q * q
	}

	mean := sum / float64(count)
	variance := (sqSum / float64(count)) - (mean * mean)
	if variance == 0 {
		return 0, mean
	}

	return math.Sqrt(variance), mean
}

func (s *Score) calculateSizeZScore(ob *OrderBook) error {
	countBids := min(LevelCount, len(ob.Bids))
	countAsks := min(LevelCount, len(ob.Asks))

	if countBids == 0 || countAsks == 0 {
		return ErrLevelCountMustBeGreaterThanZero
	}

	bidsStdDev, bidsMean := s.calculateStdDevAndMean(ob.Bids)
	asksStdDev, asksMean := s.calculateStdDevAndMean(ob.Asks)

	if bidsStdDev == 0 || asksStdDev == 0 {
		return ErrStdDevMustBeGreaterThanZero
	}

	for i := 0; i < countBids; i++ { // or countAsks
		ob.Bids[i].ZScore = (ob.Bids[i].Quantity - bidsMean) / bidsStdDev
		ob.Asks[i].ZScore = (ob.Asks[i].Quantity - asksMean) / asksStdDev
	}

	//s.print(*ob)
	return nil
}

func (s *Score) print(orderBook OrderBook) {
	fmt.Print("\033[H\033[2J")
	log.Println("Pair: ETH_TL | Outliers Detection: 1.5 | Level Count: 20")

	fmt.Println("+--------------------------------------+--------------------------------------+")
	fmt.Println("|                BIDS                  |                ASKS                  |")
	fmt.Println("+--------------------------------------+--------------------------------------+")
	fmt.Println("|   Z-Score  |   Price    |  Quantity  |    Price   |  Quantity  |   Z-Score  |")

	normal := color.New(color.FgWhite).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bg := color.New(color.BgWhite).SprintFunc()

	maxRows := len(orderBook.Bids)
	if len(orderBook.Asks) > maxRows {
		maxRows = len(orderBook.Asks)
	}

	for i := 0; i < maxRows; i++ {
		var bidStr, askStr string

		if i < len(orderBook.Bids) {
			bid := orderBook.Bids[i]
			if math.Abs(bid.ZScore) > OutliersDetection {
				// Outlier
				bidStr = bg(fmt.Sprintf("%s | %s | %s", fmt.Sprintf("%10.2f", bid.ZScore), fmt.Sprintf("%10.2f", bid.Price), fmt.Sprintf("%10.6f", bid.Quantity)))
			} else {
				bidStr = fmt.Sprintf("%s | %s | %s", normal(fmt.Sprintf("%10.2f", bid.ZScore)), green(fmt.Sprintf("%10.2f", bid.Price)), green(fmt.Sprintf("%10.6f", bid.Quantity)))
			}

		} else {
			bidStr = "                    "
		}

		if i < len(orderBook.Asks) {
			ask := orderBook.Asks[i]
			if math.Abs(ask.ZScore) > OutliersDetection {
				// Outlier
				askStr = bg(fmt.Sprintf("%s | %s | %s", fmt.Sprintf("%10.2f", ask.Price), fmt.Sprintf("%10.6f", ask.Quantity), fmt.Sprintf("%10.2f", ask.ZScore)))
			} else {
				askStr = fmt.Sprintf("%s | %s | %s", red(fmt.Sprintf("%10.2f", ask.Price)), red(fmt.Sprintf("%10.6f", ask.Quantity)), normal(fmt.Sprintf("%10.2f", ask.ZScore)))
			}

		} else {
			askStr = "                    "
		}

		fmt.Printf("| %s | %s |\n", bidStr, askStr)
	}

	fmt.Println("+--------------------------------------+--------------------------------------+")
}
