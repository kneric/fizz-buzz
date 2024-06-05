package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func SingleFizzBuzz(n int) string {
	if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz"
	} else if n%3 == 0 {
		return "Fizz"
	} else if n%5 == 0 {
		return "Buzz"
	}
	// By default, return the integer number n without any operation
	return strconv.Itoa(n)
}

func RangeFizzBuzzHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, errFrom := strconv.Atoi(fromStr)
	to, errTo := strconv.Atoi(toStr)
	// range "from" until "to" is inclusive
	numRange := to - from + 1

	if errFrom != nil || errTo != nil {
		errMsg := "Invalid parameters: input must be an integer"
		http.Error(w, errMsg, http.StatusBadRequest)
		log.Printf("Request: from:%s to:%s, Response: %v, Latency: %v\n", r.URL.Query().Get("from"), r.URL.Query().Get("to"), errMsg, time.Since(startTime))
		return
	}

	// from <= to
	if from > to {
		errMsg := "Invalid parameters: 'from' cannot be greater than 'to'"
		http.Error(w, errMsg, http.StatusBadRequest)
		log.Printf("Request: from:%s to:%s, Response: %v, Latency: %v\n", r.URL.Query().Get("from"), r.URL.Query().Get("to"), errMsg, time.Since(startTime))
		return
	}

	//Accept at maximum 100 numbers as the range
	if numRange > 100 {
		errMsg := "Invalid parameters: the maximum range from 'from' to 'to' is 100"
		http.Error(w, errMsg, http.StatusBadRequest)
		log.Printf("Request: from:%s to:%s, Response: %v, Latency: %v\n", r.URL.Query().Get("from"), r.URL.Query().Get("to"), errMsg, time.Since(startTime))
		return
	}

	// max timeout is 1 second
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	results := make([]string, numRange)
	var wg sync.WaitGroup

	// Technically can use 1000 as a maximum goroutine for the calculation at same time,
	// however it's not needed atm since the max range is 100
	calcChan := make(chan struct{}, numRange)

	for i := from; i <= to; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			calcChan <- struct{}{}
			select {
			case <-ctx.Done():
				return
			default:
				results[i-from] = SingleFizzBuzz(i)
			}
			<-calcChan
		}(i)
	}

	wg.Wait()

	// return string with space delimiter
	response := strings.Join(results, " ")
	w.Write([]byte(response))
	log.Printf("Request: from:%s to:%s, Response: %v, Latency: %v\n", r.URL.Query().Get("from"), r.URL.Query().Get("to"), response, time.Since(startTime))
}

func main() {
	http.HandleFunc("/range-fizzbuzz", RangeFizzBuzzHandler)

	server := &http.Server{Addr: ":3000"}
	log.Println("Server is running")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server is exiting")
}
