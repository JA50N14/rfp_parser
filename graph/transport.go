package graph

import (
	"time"
	"net/http"
)

type GraphListResponse[T any] struct {
	Value []T `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}

const maxRetries = 5


func (cfg *apiConfig) do[T any](req *http.Request) (T, error) {
	var zero T
	var (
		result T
		retryable bool
		wait time.Duration
		err error
	)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			//If server doesn't specify Retry-After, apply exponential backoff
			if wait == 0 {
				wait = backoff(attempt) //////////////////Create backoff function
			}
			time.Sleep(wait)
		}

		result, retryable, wait, err = cfg.doOnce[T](req)
		if err == nil {
			return result, nil
		}

		if !retryable {
			return zero, err
		}
	}

	return zero, fmt.Errorf("max retries exceeded")
}



func (cfg *apiConfig) doOnce[T any](req *http.Request) (T, bool, time.Duration, error) {
	var zero T
	var retryAfter time.Duration

	resp, err := cfg.client.Do(req)
	if err != nil {
		return zero, true, retryAfter, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 || resp.StatusCode >= 500 {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if second, _ := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return zero, true, retryAfter, fmt.Errorf("graph api error: status=%d", resp.StatusCode)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return zero, false, retrtAfter, fmt.Errorf("graph api error: status=%d body=%s", resp.StatusCode, string(body))
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
func (cfg *apiConfig) listAll[T any](req *http.Request) ([]T, error) {
	var all []T
	
	for {
		page, err := cfg.do[GraphListResponse[T]](req)
		if err != nil {
			return nil, fmt.Errorf("sending request: %w", err)
		}
		all = append(all, page.Value...)
		
		if page.NextLink == "" {
			return all, nil
		}
		nextReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, page.NextLink, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		nextReq.Header = req.Header.Clone()
		req = nextReq
	}
}
