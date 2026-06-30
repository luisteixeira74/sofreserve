package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
)

const (
	endpoint = "http://localhost:8080/events/djids3wi/checkin"
	token    = "tkt_TESTE_123"
	requests = 1000
)

type Metrics struct {
	success   int
	fail      int
	errors    int
	latencies []time.Duration

	mu sync.Mutex
}

func (m *Metrics) recordSuccess(latency time.Duration) {
	m.mu.Lock()
	m.success++
	m.latencies = append(m.latencies, latency)
	m.mu.Unlock()
}

func (m *Metrics) recordFail(latency time.Duration) {
	m.mu.Lock()
	m.fail++
	m.latencies = append(m.latencies, latency)
	m.mu.Unlock()
}

func (m *Metrics) recordError() {
	m.mu.Lock()
	m.errors++
	m.mu.Unlock()
}

func TestStressCheckinPerformance(t *testing.T) {
	startTest := time.Now()

	var wg sync.WaitGroup
	metrics := &Metrics{
		latencies: make([]time.Duration, 0, requests),
	}

	wg.Add(requests)

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()

			form := url.Values{}
			form.Add("token", token)

			start := time.Now()
			resp, err := http.PostForm(endpoint, form)
			latency := time.Since(start)

			if err != nil {
				metrics.recordError()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				metrics.recordSuccess(latency)
			} else {
				metrics.recordFail(latency)
			}
		}()
	}

	wg.Wait()

	totalTestTime := time.Since(startTest)

	// ======================
	// CALC METRICS
	// ======================
	if len(metrics.latencies) == 0 {
		t.Fatal("no latencies recorded")
	}

	var total time.Duration
	min := metrics.latencies[0]
	max := metrics.latencies[0]

	for _, l := range metrics.latencies {
		total += l
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}

	avg := total / time.Duration(len(metrics.latencies))
	rps := float64(requests) / totalTestTime.Seconds()

	// ======================
	// REPORT
	// ======================
	fmt.Println("\n=== SOFRESERVE STRESS TEST REPORT ===")
	fmt.Println("Total Requests:", requests)
	fmt.Println("Success:", metrics.success)
	fmt.Println("Fail:", metrics.fail)
	fmt.Println("Errors:", metrics.errors)

	fmt.Println("\nLatency:")
	fmt.Println("Avg:", avg)
	fmt.Println("Min:", min)
	fmt.Println("Max:", max)

	fmt.Println("\nThroughput:")
	fmt.Printf("Total Time: %v\n", totalTestTime)
	fmt.Printf("Requests/sec: %.2f\n", rps)

	fmt.Println("=====================================")

	if metrics.success != 1 {
		t.Fatalf("expected exactly 1 success due to atomic check-in, got %d", metrics.success)
	}
}