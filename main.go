package main

import (
	"sync"

	"github.com/cameron-wags/posty/endpoint"
)

func main() {
	ep := &endpoint.D{
		ResponsePfx: "TesT",
	}

	done := make(chan bool)
	responses := make(chan endpoint.EndpointResult, 1000)
	wg := &sync.WaitGroup{}

	// should run once per runner
	{
		wg.Add(1)
		go doRequests(ep, 10, responses, wg)
	}

	// if a lot of doRequests() need to get spun up, the responses
	// channel might get filled before this runs. I'm not sure
	// how to observe this happening though.
	p := endpoint.JSONPrinter{}
	go p.Collect(responses, done)

	wg.Wait()
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
