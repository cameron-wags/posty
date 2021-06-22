package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cameron-wags/posty/endpoint"
)

func main() {
	ep := endpoint.NewNet("http://localhost:8080/test")

	done := make(chan bool)
	responses := make(chan endpoint.EndpointResult, 30000) // not really sure how big this should be

	w := endpoint.NewFileWriter("result.json")
	go w.Collect(responses, done)

	totalReq := 100000      // requests to send in total
	targetRate := 3000      // target requests per second
	scheduledReq := 0       // counts requests dispatched in total, so we don't overshoot the total request number.
	var runningReq int32    // counts requests in flight
	var procLimit int32 = 1 // counts allowed requests in flight, gets adjusted every now and again
	var doneReq int32 = 0   // counts requests COMPLETED, so we don't under-report rates.

	// checks our request rate every so often, adjusting procLimit to get measured rates closer to targetRate.
	go watchdog(&procLimit, &doneReq, totalReq, float64(targetRate), 1000*time.Millisecond)

	wg := &sync.WaitGroup{}
	st := time.Now()
	for scheduledReq < totalReq {
		//TODO somehow don't spinwait
		if atomic.LoadInt32(&runningReq) < atomic.LoadInt32(&procLimit) {
			wg.Add(1) // this can be combined with runningReq as an adjustable semaphore
			scheduledReq++
			atomic.AddInt32(&runningReq, 1)
			go func(doneReq, releaser *int32, w *sync.WaitGroup) {
				wrapRequest(ep, responses)
				defer w.Done()
				atomic.AddInt32(doneReq, 1)
				atomic.AddInt32(releaser, -1)
			}(&doneReq, &runningReq, wg)
		}
	}

	wg.Wait()
	elapsed := time.Since(st)
	fmt.Printf("Requests: %d\tDuration: %v\tRate: %f req/s\n", totalReq, elapsed, float64(totalReq)/elapsed.Seconds())
	// run is over, tell the collector we're done and wait for exit
	close(responses)
	<-done
}

func watchdog(procLimit, doneReq *int32, reqAmt int, targetRate float64, interval time.Duration) {
	lastDone := atomic.LoadInt32(doneReq)
	lastTime := time.Now()
	t := time.NewTicker(interval)

	for pollT := range t.C {
		currentR := atomic.LoadInt32(doneReq)
		deltaR := float64(currentR - lastDone)
		deltaT := pollT.Sub(lastTime).Seconds()
		rate := deltaR / deltaT

		// dumb control strat incoming.
		// It's actually not as awful as I thought it would be.
		var intRate int32
		rateMissFactor := targetRate / rate

		newLim := float64(atomic.LoadInt32(procLimit)) * rateMissFactor
		intRate = int32(newLim)
		atomic.StoreInt32(procLimit, intRate)

		fmt.Printf("\nRate: %.2f req/s\tMaxProcs: %d\n", rate, intRate)
		fmt.Printf("Run: %d/%d - %.2f%%\n", currentR, reqAmt, float32(currentR*100)/float32(reqAmt))

		lastDone = currentR
		lastTime = pollT
	}
}

func wrapRequest(ep endpoint.EndpointSender, resultChan chan<- endpoint.EndpointResult) {
	resultChan <- ep.CreateSend()
}
