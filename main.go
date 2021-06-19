package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/cameron-wags/posty/endpoint"
)

func main() {
	concurrency := 6000
	requestPerRunner := 100

	ep := endpoint.NewNet("http://localhost:8080/test")

	done := make(chan bool)
	responses := make(chan endpoint.EndpointResult, 1000)
	wg := &sync.WaitGroup{}

	st := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go doRequests(ep, requestPerRunner, responses, wg)
	}

	// if a lot of doRequests() need to get spun up, the responses
	// channel might get filled before this runs. I'm not sure
	// how to observe this happening though.
	w := endpoint.NewFileWriter("result.json")
	go w.Collect(responses, done)

	wg.Wait()
	elapsed := time.Since(st)
	fmt.Printf("Requests: %d\tDuration: %v\tRate: %f req/s\n", concurrency*requestPerRunner, elapsed, float64(concurrency*requestPerRunner)/elapsed.Seconds())
	// run is over, tell the collector we're done and wait for exit
	close(responses)
	<-done
}

func doRequests(ep endpoint.EndpointSender, iterations int, resultChan chan<- endpoint.EndpointResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for it := 0; it < iterations; it++ {
		resultChan <- ep.CreateSend()
	}
}
