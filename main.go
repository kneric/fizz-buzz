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
	return strconv.Itoa(n)
}

func RangeFizzBuzzHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, errFrom := strconv.Atoi(fromStr)
	to, errTo := strconv.Atoi(toStr)

	if errFrom != nil || errTo != nil || from > to || to-from > 100 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		log.Printf("Request: %s, Response: Invalid parameters, Latency: %v\n", r.URL.String(), time.Since(startTime))
		return
	}

	// max timeout is 1 second
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	// iterate "from" until "to" and inclusive
	results := make([]string, to-from+1)
	var wg sync.WaitGroup

	calcChan := make(chan struct{}, 1000)

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
	log.Printf("Request: %s, Response: %s, Latency: %v\n", r.URL.String(), response, time.Since(startTime))
}

func main() {
	http.HandleFunc("/range-fizzbuzz", RangeFizzBuzzHandler)

	server := &http.Server{Addr: ":3000"}

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

	log.Println("Server exiting")
}
