package main

import (
	"context"
	"github.com/starbase-343/ferengi/utils/multiplexer"
	"github.com/starbase-343/ferengi/utils/streamer/prb"
	"github.com/subspace-343/z-score/printer"
	"github.com/subspace-343/z-score/score"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	prbTicker = "eth_tl"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	prbStreamer, err := prb.NewPrbStreamer(prbTicker)
	if err != nil {
		log.Fatal(err)
	}
	go prbStreamer.AsyncRun(ctx)

	streamerCount := 1

	mpx := multiplexer.NewV1()
	if err := mpx.Attach(prbStreamer, ctx); err != nil {
		log.Fatal(err)
	}

	pr := printer.NewConsolePrinter()

	zScore := score.NewScore(pr)
	go zScore.AsyncRun(ctx, mpx, streamerCount)

	select {
	case <-signalChan:
		cancel()
		log.Println("Received shutdown signal, exiting...")

		os.Exit(0)
	}
}
