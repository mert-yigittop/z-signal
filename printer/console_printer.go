package printer

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/subspace-343/z-score/score"
	"log"
	"os"
	"os/exec"
	"runtime"
)

type ConsolePrinter struct{}

func NewConsolePrinter() Printer {
	return &ConsolePrinter{}
}

// PRINT
func (s *ConsolePrinter) clearConsole() {
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

func (s *ConsolePrinter) Print(orderBook score.OrderBook) {
	s.clearConsole()
	log.Println("INJ_TL")

	fmt.Println("+-------------------------+-------------------------+---------------+")
	fmt.Println("|          BIDS           |           ASKS          |     Z-Score   |")
	fmt.Println("+-------------------------+-------------------------+---------------+")
	fmt.Println("|      Price | Quantity   |      Price | Quantity   |               |")

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	maxRows := len(orderBook.Bids)
	if len(orderBook.Asks) > maxRows {
		maxRows = len(orderBook.Asks)
	}

	for i := 0; i < maxRows; i++ {
		var bidStr, askStr string

		if i < len(orderBook.Bids) {
			bid := orderBook.Bids[i]
			bidStr = fmt.Sprintf("%s | %s", green(fmt.Sprintf("%10.2f", bid.Price)), green(fmt.Sprintf("%10.6f", bid.Quantity)))
		} else {
			bidStr = "                    "
		}

		if i < len(orderBook.Asks) {
			ask := orderBook.Asks[i]
			askStr = fmt.Sprintf("%s | %s", red(fmt.Sprintf("%10.2f", ask.Price)), red(fmt.Sprintf("%10.6f", ask.Quantity)))
		} else {
			askStr = "                    "
		}

		fmt.Printf("| %s | %s |     %6.2f    |\n", bidStr, askStr, 0.00)
	}

	fmt.Println("+-------------------------+-------------------------+---------------+")
}
