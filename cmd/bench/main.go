package main

import (
	"flag"
	"fmt"

	"sof-reserve/internal/bench"
)

var (
	endpointFlag    = flag.String("endpoint", "", "")
	tokenFlag       = flag.String("token", "", "")
	requestsFlag    = flag.Int("requests", 0, "")
	concurrencyFlag = flag.Int("concurrency", 0, "")
)

func main() {
	flag.Parse()

	cfg := bench.DefaultConfig()

	if *endpointFlag != "" {
		cfg.Endpoint = *endpointFlag
	}
	if *tokenFlag != "" {
		cfg.Token = *tokenFlag
	}
	if *requestsFlag != 0 {
		cfg.Requests = *requestsFlag
	}
	if *concurrencyFlag != 0 {
		cfg.Concurrency = *concurrencyFlag
	}

	fmt.Println("Running benchmark on:", cfg.Endpoint)

	report := bench.Run(cfg)

	bench.Print(cfg, report)
}