package printer

import (
	"github.com/subspace-343/z-score/order"
)

type Printer interface {
	Print(orderBook order.Book)
}
