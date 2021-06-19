package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Net struct {
	URL    string
	Client *http.Client
}

func NewNet(url string) *Net {
	return &Net{
		URL: url,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Net) CreateSend() EndpointResult {
	reqBody := strings.NewReader(uuid.NewString())

	start := time.Now()
	resp, err := n.Client.Post(n.URL, "text/plain", reqBody)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return &NetResult{
			Error: err,
		}
	}

	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &NetResult{
			Error: err,
		}
	}

	return &NetResult{
		Status:     resp.StatusCode,
		Body:       string(respBytes),
		DurationMS: int(duration),
	}
}

type NetResult struct {
	Status     int
	Body       string
	DurationMS int
	Error      error
}

func (r *NetResult) Summarize() interface{} {
	return r
}

type FileWriter struct {
	Target *os.File
}

func NewFileWriter(path string) ResultCollector {
	f, e := os.Create(path)
	if e != nil {
		panic(e)
	}
	return &FileWriter{
		Target: f,
	}
}

func (f *FileWriter) Collect(results <-chan EndpointResult, done chan<- bool) {
	f.Target.WriteString("[\n")

	enc := json.NewEncoder(f.Target)
	enc.SetIndent("  ", "  ")
	first := true
	for {
		res, ok := <-results
		if !ok {
			f.Target.WriteString("]\n")
			done <- true
			return
		}
		if first {
			first = false
		} else {
			f.Target.WriteString(",")
		}
		enc.Encode(res.Summarize())
	}
}
