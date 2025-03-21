package order

type Level struct {
	Price, Quantity, ZScore float64
}

type Book struct {
	Asks []Level
	Bids []Level
}
