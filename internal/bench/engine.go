package bench

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	Duration  time.Duration
	StartedAt time.Time
}

func DefaultConfig() Config {
	return Config{
		Endpoint:    "http://localhost:8080/events/CHANGE_ME/checkin",
		Token:       "",
		Requests:    100,
		Concurrency: 10,
	}
}

func Run(cfg Config) Report {

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
	report.Preflight.Valid = true

	return report
}

func execute(cfg Config, start time.Time) Report {

	var success int64
	var fail int64
	var errors int64

	var totalLatency int64

	minLatency := time.Hour
	var maxLatency time.Duration

	var latencyMu sync.Mutex

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
			defer func() {
				<-sem
			}()

			form := url.Values{}
			form.Set("token", cfg.Token)

			req, err := http.NewRequest(
				http.MethodPost,
				cfg.Endpoint,
				strings.NewReader(form.Encode()),
			)

			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}

			req.Header.Set(
				"Content-Type",
				"application/x-www-form-urlencoded",
			)

			reqStart := time.Now()

			resp, err := client.Do(req)

			latency := time.Since(reqStart)

			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}

			defer resp.Body.Close()

			atomic.AddInt64(&totalLatency, int64(latency))

			latencyMu.Lock()

			if latency < minLatency {
				minLatency = latency
			}

			if latency > maxLatency {
				maxLatency = latency
			}

			latencyMu.Unlock()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				atomic.AddInt64(&success, 1)
			} else {
				atomic.AddInt64(&fail, 1)
			}

		}()
	}

	wg.Wait()

	duration := time.Since(start)

	avgLatency := time.Duration(
		totalLatency / int64(cfg.Requests),
	)

	rps := float64(cfg.Requests) / duration.Seconds()

	return Report{
		Requests: cfg.Requests,

		Success: int(success),
		Fail:    int(fail),
		Errors:  int(errors),

		AvgLatency: avgLatency,
		MinLatency: minLatency,
		MaxLatency: maxLatency,
		RPS:        rps,

		Duration:  duration,
		StartedAt: start,
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

	form := url.Values{}
	form.Set("token", cfg.Token)

	req, err := http.NewRequest(
		http.MethodPost,
		cfg.Endpoint,
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	start := time.Now()

	resp, err := client.Do(req)

	latency := time.Since(start)

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

	fmt.Printf("Preflight OK (%d) - %v\n", resp.StatusCode, latency)

	return nil
}