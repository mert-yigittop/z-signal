package printer

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/subspace-343/z-score/order"
	"log"
	"math"
)

var (
	// OutliersDetection is the threshold for the z-score to detect outliers
	OutliersDetection = 1.5
)

type ConsolePrinter struct{}

func NewConsolePrinter() Printer {
	return &ConsolePrinter{}
}

func (p *ConsolePrinter) Print(orderBook order.Book) {
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
