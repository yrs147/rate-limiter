package main

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"
)

type TocketBucket struct {
	mu             sync.Mutex
	capacity       int
	tokens         int
	refillRate     int
	lastRefillTime time.Time
}

// Initialize new bucket
func NewTokenBucket(capacity int, refillRate int) *TocketBucket {
	return &TocketBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Refill the buclet with tokens
func RefillBucket(bucket *TocketBucket) {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	elapsedTime := time.Since(bucket.lastRefillTime).Seconds()
	tokensToAdd := int(elapsedTime) * bucket.refillRate
	bucket.tokens = int(math.Min(float64(bucket.capacity), float64(bucket.tokens+tokensToAdd)))
	bucket.lastRefillTime = time.Now()
}

func allowRequest(bucket *TocketBucket, tokensReq int) bool {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	RefillBucket(bucket)

	//Token check
	if bucket.tokens >= tokensReq {
		bucket.tokens -= tokensReq
		return true
	}

	return false
}

func apiHandler(rw http.ResponseWriter, req *http.Request, bucket *TocketBucket) {
	if allowRequest(bucket, 1) {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintf(rw, "Success! Request processed .")
	} else {
		rw.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(rw, "Error: Too Many Requests .")
	}
}

func main() {

	bucket := NewTokenBucket(10, 1)

	apiHandlerFunc := func(w http.ResponseWriter, r *http.Request){
		apiHandler(w,r,bucket)
	}

	http.HandleFunc("/api",apiHandlerFunc)
	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080",nil)

}
