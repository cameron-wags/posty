package endpoint

// EndpointSender can create a request to send to an endpoint
// and send it. The response is returned.
type EndpointSender interface {
	// CreateSend creates whatever data is needed to send a request to
	// an endpoint, then sends it. The result of this request is returned.
	//
	// This call must block until the send is complete and the result is
	// ready for interpretation.
	CreateSend() EndpointResult
}

// EndpointResult is a set of observations taken from a request
// to an endpoint.
type EndpointResult interface {
	// Summarize returns a struct with information a tester may care about.
	// These fields are used as the result of an endpoint Send.
	// If something went wrong, this is the place to say it.
	//
	// Sample fields could be HTTPStatusCode, HTTPResponseTime, etc.
	Summarize() interface{}
}

type ResultCollector interface {
	// Collect registers an interperter of endpoint Results.
	// Collect should interpret from results until the channel is closed.
	//
	// When the results channel is closed, Collect must send a signal
	// on the done channel.
	Collect(results <-chan EndpointResult, done chan<- bool)
}
