package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func TestRaceCondition(t *testing.T) {
	token := "tkt_TESTE_123"
	endpoint := "http://localhost:8080/events/djids3wi/checkin"

	var wg sync.WaitGroup

	success := 0
	fail := 0
	errors := 0

	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			form := url.Values{}
			form.Add("token", token)

			resp, err := http.PostForm(endpoint, form)
			if err != nil {
				mu.Lock()
				errors++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			switch resp.StatusCode {
			case 200:
				success++
			default:
				fail++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println("SUCCESS:", success)
	fmt.Println("FAIL:", fail)
	fmt.Println("ERRORS:", errors)

	if success != 1 {
		t.Fatalf("expected 1 success, got %d", success)
	}
}