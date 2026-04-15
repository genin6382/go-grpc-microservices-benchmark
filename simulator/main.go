package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Generator interface {
	Next() RequestSpec
	Delay() time.Duration
}

type RequestSpec struct {
	Method string
	Path   string
	Body   string
	UserID string
}

type Result struct {
	Timestamp string
	Strategy  string
	Pattern   string
	Iter      int
	Path      string
	UserID    string
	Status    int
	LatencyMS int64
	Error     string
}

func main() {
	pattern := flag.String("pattern", "uniform", "uniform | bursty | skewed")
	strategy := flag.String("strategy", "round_robin", "label for output")
	baseURL := flag.String("base", "http://localhost:8000", "gateway base url")
	total := flag.Int("n", 100, "number of requests")
	concurrency := flag.Int("c", 10, "concurrency")
	out := flag.String("out", "output/results.csv", "output csv")
	appendMode := flag.Bool("append", false, "append to existing csv")
	flag.Parse()

	var gen Generator
	switch *pattern {
	case "uniform":
		gen = NewUniform()
	case "bursty":
		gen = NewBursty()
	case "skewed":
		gen = NewSkewed()
	default:
		panic("invalid pattern")
	}

	if err := os.MkdirAll("output", 0o755); err != nil {
		panic(err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	jobs := make(chan int)
	results := make(chan Result, *total)

	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				spec := gen.Next()
				url := strings.TrimRight(*baseURL, "/") + spec.Path

				req, err := http.NewRequest(spec.Method, url, strings.NewReader(spec.Body))
				if err != nil {
					results <- Result{
						Timestamp: time.Now().Format(time.RFC3339),
						Strategy:  *strategy,
						Pattern:   *pattern,
						Iter:      idx,
						Path:      spec.Path,
						UserID:    spec.UserID,
						Error:     err.Error(),
					}
					continue
				}

				if spec.Body != "" {
					req.Header.Set("Content-Type", "application/json")
				}
				if spec.UserID != "" {
					req.Header.Set("X-User-ID", spec.UserID)
				}

				start := time.Now()
				resp, err := client.Do(req)
				latency := time.Since(start).Milliseconds()

				if err != nil {
					results <- Result{
						Timestamp: time.Now().Format(time.RFC3339),
						Strategy:  *strategy,
						Pattern:   *pattern,
						Iter:      idx,
						Path:      spec.Path,
						UserID:    spec.UserID,
						LatencyMS: latency,
						Error:     err.Error(),
					}
					continue
				}

				resp.Body.Close()

				results <- Result{
					Timestamp: time.Now().Format(time.RFC3339),
					Strategy:  *strategy,
					Pattern:   *pattern,
					Iter:      idx,
					Path:      spec.Path,
					UserID:    spec.UserID,
					Status:    resp.StatusCode,
					LatencyMS: latency,
				}
			}
		}()
	}

	go func() {
		for i := 0; i < *total; i++ {
			jobs <- i
			time.Sleep(gen.Delay())
		}
		close(jobs)
	}()

	wg.Wait()
	close(results)

	writeResults(*out, *appendMode, results)
}

func writeResults(path string, appendMode bool, results <-chan Result) {
	var f *os.File
	var err error

	if appendMode {
		f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	} else {
		f, err = os.Create(path)
	}
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	info, _ := f.Stat()
	needHeader := !appendMode || info.Size() == 0
	if needHeader {
		_ = w.Write([]string{
			"timestamp", "strategy", "pattern", "iter", "path", "user_id", "status", "latency_ms", "error",
		})
	}

	for r := range results {
		_ = w.Write([]string{
			r.Timestamp,
			r.Strategy,
			r.Pattern,
			fmt.Sprint(r.Iter),
			r.Path,
			r.UserID,
			fmt.Sprint(r.Status),
			fmt.Sprint(r.LatencyMS),
			r.Error,
		})
	}
}