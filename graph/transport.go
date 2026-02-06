package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/JA50N14/rfp_parser/config"
)

type GraphListResponse[T any] struct {
	Value []T `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}

const maxRetries = 5


func do[T any](ctx context.Context, cfg *config.ApiConfig, buildReq func(ctx context.Context) (*http.Request, error)) (T, error) {
	var zero T

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := buildReq(ctx)
		if err != nil {
			return zero, fmt.Errorf("build request: %w", err)
		}

		result, retryable, wait, err := doOnce[T](req, cfg)
		if err == nil {
			return result, nil
		}

		if !retryable {
			return zero, err
		}

		//exponential backoff if server does not provide Retry-After
		if wait == 0 {
			wait = backoff(attempt + 1)
		}

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return zero, fmt.Errorf("context error: %w", ctx.Err())
		}
	}

	return zero, fmt.Errorf("max retries exceeded")
}

func doOnce[T any](req *http.Request, cfg *config.ApiConfig) (T, bool, time.Duration, error) {
	var zero T
	var retryAfter time.Duration

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return zero, true, retryAfter, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 || resp.StatusCode >= 500 {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, err := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return zero, true, retryAfter, fmt.Errorf("graph api error: status=%d", resp.StatusCode)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return zero, false, retryAfter, fmt.Errorf("graph api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result T
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		return zero, false, retryAfter, fmt.Errorf("decode response: %w", err)
	}
	
	return result, false, retryAfter, nil
}  


//Pagination
func listAll[T any](ctx context.Context, cfg *config.ApiConfig, buildReq func(ctx context.Context) (*http.Request, error)) ([]T, error) {
	var all []T

	for {
		page, err := do[GraphListResponse[T]](ctx, cfg, buildReq)
		if err != nil {
			return nil, fmt.Errorf("sending request: %w", err)
		}
		all = append(all, page.Value...)
		
		if page.NextLink == "" {
			return all, nil
		}

		nextLink := page.NextLink
		prevBuild := buildReq 

		buildReq = func(ctx context.Context) (*http.Request, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextLink, nil)
			if err != nil {
				return nil, err
			}
			prevReq, err := prevBuild(ctx)
			if err == nil {
				req.Header = prevReq.Header.Clone()
			}
			return req, nil
		}
	}
}

func backoff(attempt int) time.Duration {
	base := time.Second
	max := 30 * time.Second

	d := time.Duration(1<<attempt) * base
	if d > max {
		d = max
	}

	jitter := time.Duration(rand.Int64N(int64(d / 2)))

	return d/2 + jitter
}





