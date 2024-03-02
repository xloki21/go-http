package source

import (
	"context"
	"github.com/xloki21/go-http/internal/model"
	"io"
	"net/http"
	"sync"
	"time"
)

func FetchURLList(ctx context.Context, urls []model.URL, maxConcurrentRequests int, timeoutPerRequest time.Duration) ([]model.Result, error) {
	var wg sync.WaitGroup
	tasks := make(chan model.URL, len(urls))
	errors := make(chan error, len(urls))
	resChan := make(chan model.Result, len(urls))

	for w := 0; w < maxConcurrentRequests; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				url, ok := <-tasks
				if !ok {
					return
				}

				body, err := fetchURLWithTimeout(ctx, string(url), timeoutPerRequest)
				if err != nil {
					errors <- err
					return
				}
				resChan <- model.Result{URL: url, Content: body}
			}
		}()
	}

	for _, url := range urls {
		tasks <- url
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(resChan)
		close(errors)
	}()

	var err error
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	var results []model.Result
	for result := range resChan {
		results = append(results, result)
	}

	return results, err
}

func fetchURLWithTimeout(ctx context.Context, url string, timeout time.Duration) ([]byte, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
