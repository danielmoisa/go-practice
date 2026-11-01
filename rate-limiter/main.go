package main

import (
	"fmt"
	"time"
)

// RateLimiter controls request throughput
type RateLimiter struct {
	requestsPerSecond int
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		requestsPerSecond,
	}
}

func (rl *RateLimiter) Process(requests <-chan int, results chan<- string) {
	// Calculate the interval between requests
	interval := time.Second / time.Duration(rl.requestsPerSecond)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case req, ok := <-requests:
			if !ok {
				// Channel closed, graceful shutdown
				return
			}
			// Wait for the ticker to allow the next request
			<-ticker.C
			results <- fmt.Sprintf("Processed request %d", req)
		}
	}
}

func main() {
	rl := NewRateLimiter(10)

	requests := make(chan int)
	results := make(chan string)

	go rl.Process(requests, results)

	go func() {
		for i := 1; i <= 15; i++ {
			requests <- i
		}
		close(requests) // Signal we're done
	}()

	for result := range results {
		fmt.Println(result)
	}
}
