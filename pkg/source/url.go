package source

import (
	"context"
	"github.com/xloki21/go-http/internal/model"
	"io"
	"net/http"
	"net/url"
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
				urlAddr, ok := <-tasks
				if !ok {
					return
				}

				body, err := fetchURLWithTimeout(ctx, string(urlAddr), timeoutPerRequest)
				if err != nil {
					errors <- err
					return
				}
				resChan <- model.Result{URL: urlAddr, Content: body}
			}
		}()
	}

	for _, urlAddr := range urls {
		tasks <- urlAddr
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(resChan)
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	results := make([]model.Result, 0, len(urls))
	for result := range resChan {
		results = append(results, result)
	}

	return results, nil
}

func fetchURLWithTimeout(ctx context.Context, urlString string, timeout time.Duration) ([]byte, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if _, err := url.Parse(urlString); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
