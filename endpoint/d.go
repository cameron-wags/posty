package endpoint

/*
The following is a rushed implementation of the endpoint interface used in testing.
*/

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type D struct {
	ResponsePfx string
}

type DR struct {
	Response string
}

func (this *D) CreateSend() EndpointResult {
	v := &DR{
		Response: fmt.Sprintf("%s-%s", this.ResponsePfx, uuid.NewString()),
	}
	return v
}

func (this *DR) Summarize() interface{} {
	return this
}

type JSONPrinter struct {
}

func (this *JSONPrinter) Collect(results <-chan EndpointResult, done chan<- bool) {
	for {
		res, ok := <-results
		if !ok {
			done <- true
			return
		}
		b, _ := json.Marshal(res.Summarize())
		fmt.Printf("%s\n", b)
	}
}
