package walk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/JA50N14/rfp_parser/config"
)

const maxRetries = 5

func postToSmartsheets(smartsheetRows []Row, ctx context.Context, cfg *config.ApiConfig) error {
	payloadBytes, err := json.Marshal(smartsheetRows)
	if err != nil {
		return err
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.SmartsheetUrl, bytes.NewReader(payloadBytes))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.BearerTokenSmartsheet)

		resp, err := cfg.Client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				if err := backoff(ctx, attempt); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("smartsheet request failed after %d attempts: %w", attempt, err)
		}

		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}

		//Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		//Determine retryability
		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			lastErr = fmt.Errorf("rate limited: %s", string(bodyBytes))

		case resp.StatusCode == http.StatusRequestTimeout:
			lastErr = fmt.Errorf("request timeout: %s", string(bodyBytes))

		case resp.StatusCode >= 500 && resp.StatusCode <= 599:
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(bodyBytes))

		default:
			//Non-retryable client error
			return fmt.Errorf("non-retryable smartsheet error %d: %s", resp.StatusCode, string(bodyBytes))
		}

		//Retry if attempts remain
		if attempt < maxRetries {
			if err := backoff(ctx, attempt); err != nil {
				return err
			}
			continue
		}

		return fmt.Errorf("smartsheet request failed after %d attempts: %w", attempt, lastErr)
	}

	return lastErr
}

func backoff(ctx context.Context, attempt int) error {
	base := time.Second
	max := 30 * time.Second

	d := time.Duration(1<<attempt) * base
	if d > max {
		d = max
	}

	jitter := time.Duration(rand.Int64N(int64(d / 2)))
	sleep := d/2 + jitter

	timer := time.NewTimer(sleep)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
