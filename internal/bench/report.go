package bench

import "fmt"

func Print(cfg Config, r Report) {

	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("        SOFRESERVE HTTP BENCH")
	fmt.Println("========================================")

	fmt.Printf("Endpoint:      %s\n", cfg.Endpoint)
	fmt.Printf("Started At:    %s\n", r.StartedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Concurrency:   %d\n", cfg.Concurrency)
	fmt.Printf("Requests:      %d\n", cfg.Requests)

	fmt.Println("----------------------------------------")

	fmt.Printf("Success:       %d\n", r.Success)
	fmt.Printf("Fail:          %d\n", r.Fail)
	fmt.Printf("Errors:        %d\n", r.Errors)

	fmt.Println("----------------------------------------")

	fmt.Printf("Average:       %v\n", r.AvgLatency)
	fmt.Printf("Minimum:       %v\n", r.MinLatency)
	fmt.Printf("Maximum:       %v\n", r.MaxLatency)

	fmt.Println("----------------------------------------")

	fmt.Printf("RPS:           %.2f\n", r.RPS)
	fmt.Printf("Duration:      %v\n", r.Duration)

	fmt.Println("========================================")
}