package bench

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	Endpoint    string
	Token       string
	Requests    int
	Concurrency int
}

type PreflightResult struct {
	StatusCode int
	Latency    time.Duration
	Valid      bool
	Error      string
}

type Report struct {
	Requests   int
	Success    int
	Fail       int
	Errors     int
	AvgLatency time.Duration
	MinLatency time.Duration
	MaxLatency time.Duration
	RPS        float64

	Preflight PreflightResult
	Duration   time.Duration
	StartedAt  time.Time
}

func DefaultConfig() Config {
	return Config{
		Endpoint:    "http://localhost:8080/events/djids3wi/checkin",
		Token:       "tkt_TESTE_123",
		Requests:    1000,
		Concurrency: 100,
	}
}

func Run(cfg Config) Report {
	fmt.Println("Running benchmark on:", cfg.Endpoint)

	start := time.Now()

	if err := validate(cfg); err != nil {
		fmt.Println("config error:", err)
		return Report{}
	}

	if err := preflight(cfg); err != nil {
		fmt.Println("preflight error:", err)
		return Report{}
	}

	report := execute(cfg, start)

	report.Preflight = PreflightResult{
		Valid: true,
	}

	return report
}

func execute(cfg Config, start time.Time) Report {
	var success int64
	var fail int64
	var errors int64

	var totalLatency int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.Concurrency)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < cfg.Requests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			reqBody := []byte(fmt.Sprintf(`{"token":"%s"}`, cfg.Token))

			req, err := http.NewRequest("POST", cfg.Endpoint, bytes.NewBuffer(reqBody))
			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			reqStart := time.Now()
			resp, err := client.Do(req)
			latency := time.Since(reqStart)

			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}
			defer resp.Body.Close()

			atomic.AddInt64(&totalLatency, int64(latency))

			if latency < time.Duration(minLatency) {
				minLatency = int64(latency)
			}
			if latency > time.Duration(maxLatency) {
				maxLatency = int64(latency)
			}

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				atomic.AddInt64(&success, 1)
			} else {
				atomic.AddInt64(&fail, 1)
			}
		}()
	}

	wg.Wait()

	duration := time.Since(start)

	//total := float64(cfg.Requests)
	avgLatency := time.Duration(totalLatency / int64(cfg.Requests))
	rps := float64(cfg.Requests) / duration.Seconds()

	return Report{
		Requests:   cfg.Requests,
		Success:    int(success),
		Fail:       int(fail),
		Errors:     int(errors),
		AvgLatency: avgLatency,
		MinLatency: time.Duration(minLatency),
		MaxLatency: time.Duration(maxLatency),
		RPS:        rps,
		Duration:   duration,
		StartedAt:  start,
	}
}

func validate(cfg Config) error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if cfg.Token == "" {
		return fmt.Errorf("token is required")
	}

	if cfg.Requests <= 0 {
		return fmt.Errorf("requests must be > 0")
	}

	if cfg.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be > 0")
	}

	return nil
}

func preflight(cfg Config) error {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("POST", cfg.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// ajuste aqui conforme seu sistema (header/token)
	req.Header.Set("Authorization", cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("endpoint unreachable: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("event or token not found (404)")
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid token (401)")
	default:
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error before benchmark: %d", resp.StatusCode)
		}
	}

	return nil
}