package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/healthz", "URL to test")
	concurrency := flag.Int("c", 10, "concurrent workers")
	total := flag.Int("n", 1000, "total requests")
	flag.Parse()

	var (
		successCount int64
		failCount    int64
		totalLatency int64 // microseconds
		minLatency   int64 = 1<<63 - 1
		maxLatency   int64
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        *concurrency,
			MaxIdleConnsPerHost: *concurrency,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	var wg sync.WaitGroup
	requests := make(chan int, *total)
	for i := 0; i < *total; i++ {
		requests <- i
	}
	close(requests)

	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requests {
				reqStart := time.Now()
				resp, err := client.Get(*url)
				latency := time.Since(reqStart).Microseconds()
				if err != nil {
					atomic.AddInt64(&failCount, 1)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failCount, 1)
				}
				atomic.AddInt64(&totalLatency, latency)
				// Update min/max using atomic operations for simplicity
				for {
					old := atomic.LoadInt64(&minLatency)
					if latency >= old || atomic.CompareAndSwapInt64(&minLatency, old, latency) {
						break
					}
				}
				for {
					old := atomic.LoadInt64(&maxLatency)
					if latency <= old || atomic.CompareAndSwapInt64(&maxLatency, old, latency) {
						break
					}
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("Load test completed\n")
	fmt.Printf("URL: %s\n", *url)
	fmt.Printf("Total requests: %d\n", *total)
	fmt.Printf("Concurrency: %d\n", *concurrency)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Success: %d\n", atomic.LoadInt64(&successCount))
	fmt.Printf("Failed: %d\n", atomic.LoadInt64(&failCount))
	if *total > 0 {
		avgLatency := float64(atomic.LoadInt64(&totalLatency)) / float64(*total)
		fmt.Printf("Average latency: %.2f ms\n", avgLatency/1000)
		fmt.Printf("Min latency: %.2f ms\n", float64(atomic.LoadInt64(&minLatency))/1000)
		fmt.Printf("Max latency: %.2f ms\n", float64(atomic.LoadInt64(&maxLatency))/1000)
		fmt.Printf("Requests/sec: %.2f\n", float64(*total)/duration.Seconds())
	}
}
