package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fabiomigueldp/ante/internal/sim"
)

func main() {
	hands := flag.Int("hands", 1000, "number of random hands to simulate")
	seed := flag.Int64("seed", 42, "random seed")
	flag.Parse()

	report := sim.RunRandomHands(*hands, *seed)
	fmt.Fprintf(os.Stdout, "simulated=%d seeds=%d panics=%d\n", report.HandsSimulated, report.SeedsTested, report.Panics)
	if report.Panics > 0 {
		os.Exit(1)
	}
}
