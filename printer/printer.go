package printer

import "github.com/subspace-343/z-score/score"

type Printer interface {
	Print(book score.OrderBook)
}
