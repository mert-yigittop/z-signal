package score

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/starbase-343/ferengi/utils/multiplexer"
	"github.com/starbase-343/ferengi/utils/streamer"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
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
	if err := s.calculateBidsSizeZScore(&ob); err != nil {
		return err
	}

	return nil
}

func (s *Score) calculateBidsSizeMean(ob OrderBook) float64 {
	var sum float64
	count := min(LevelCount, len(ob.Bids))

	// If there are no bids, return 0
	if count == 0 {
		return 0
	}

	for i := 0; i < count; i++ {
		sum += ob.Bids[i].Quantity
	}

	return sum / float64(count)
}

func (s *Score) calculateBidsSizeStdDev(ob OrderBook) (float64, float64) {
	mean := s.calculateBidsSizeMean(ob)

	var sum float64
	count := min(LevelCount, len(ob.Bids))

	// If there are no bids, return 0
	if count == 0 {
		return 0, mean
	}

	for i := 0; i < count; i++ {
		diff := ob.Bids[i].Quantity - mean // X_i - μ_SPB
		sum += diff * diff                 // (X_i - μ_SPB)^2
	}

	return math.Sqrt(sum / float64(count)), mean
}

func (s *Score) calculateBidsSizeZScore(ob *OrderBook) error {
	//mean := s.calculateBidsSizeMean(*ob)
	stdDev, mean := s.calculateBidsSizeStdDev(*ob)

	count := min(LevelCount, len(ob.Bids))

	if count == 0 {
		return ErrLevelCountMustBeGreaterThanZero
	}

	if stdDev == 0 {
		return ErrStdDevMustBeGreaterThanZero
	}

	for i := 0; i < count; i++ {
		z := (ob.Bids[i].Quantity - mean) / stdDev // Z-score = (X_i - μ_SPB) / σ_SPB
		ob.Bids[i].ZScore = z
	}

	s.print(*ob)
	return nil
}

// PRINT
func (s *Score) clearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func (s *Score) print(orderBook OrderBook) {
	s.clearConsole()
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
			askStr = fmt.Sprintf("%s | %s | %s", red(fmt.Sprintf("%10.2f", ask.Price)), red(fmt.Sprintf("%10.6f", ask.Quantity)), normal(fmt.Sprintf("%10.2f", ask.ZScore))) // TODO: Change asks zscore
		} else {
			askStr = "                    "
		}

		fmt.Printf("| %s | %s |\n", bidStr, askStr)
	}

	fmt.Println("+--------------------------------------+--------------------------------------+")
}
