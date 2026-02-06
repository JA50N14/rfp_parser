package graph

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/JA50N14/rfp_parser/config"
)

const maxChunks = 10_000

func GetFile(ctx context.Context, cfg *config.ApiConfig, itemID string) (*os.File, error) {
	var written int64
	var totalSize int64 = -1
	var chunks int64 = 0
	retries := 0

	tmp, err := os.CreateTemp("", "temp*")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}

	success := false
	defer func() {
		if !success {
			os.Remove(tmp.Name())
		}
	}()
	
	for {
		req, err := CreateGetFileRequest(ctx, cfg, itemID, written)
		if err != nil {
			return nil, err
		}
		
		resp, err := cfg.Client.Do(req)
		if err != nil {
			if retries >= maxRetries {
				return nil, fmt.Errorf("maximum retries exceeded")
			}
			if ctx.Err() != nil {
				return nil, fmt.Errorf("context error: %w", ctx.Err())
			}
			retries++
			continue
		}

		var copyErr error
		fatal := false

		func() {
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				if written > 0 {
					err = fmt.Errorf("server ignored range request")
					fatal = true
					return
				}
				var n int64
				if _, err = tmp.Seek(written, io.SeekStart); err != nil {
					fatal = true
					return
				}
				
				if cl := resp.ContentLength; cl >= 0 {
					totalSize = cl
				}

				n, copyErr = io.Copy(tmp, resp.Body)
				written += n

			case http.StatusPartialContent:
				var n int64
				if _, err = tmp.Seek(written, io.SeekStart); err != nil {
					fatal = true
					return
				}

				rangeHeader := resp.Header.Get("Content-Range")
				var start, end, total int64
				start, end, total, err = partialContentPreCopyCheckSum(rangeHeader, written)
				if err != nil {
					fatal = true
					return
				}

				if totalSize == -1 {
					totalSize = total
				}

				if totalSize != -1 && total != totalSize {
					err = fmt.Errorf("total size changed from %d to %d", totalSize, total)
					fatal = true
					return
				}


				n, copyErr = io.Copy(tmp, resp.Body)
				written += n

				if chunks >= maxChunks || n <= 0 {
					err = fmt.Errorf("lack of progress made on partial content: %d chunks", chunks)
					fatal = true
					return
				}

				if n != (end - start + 1) {
					err = fmt.Errorf("copied %d bytes, expected %d", n, end - start + 1)
					fatal = true
					return
				}
				chunks++
				retries = 0
				
			case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
				err = fmt.Errorf("download failed: %s", resp.Status)
				fatal = true

			case http.StatusRequestedRangeNotSatisfiable:
				if totalSize != -1 && written >= totalSize {
					return
				}
				err = fmt.Errorf("server error: %d", written)
				fatal = true
			
			case http.StatusTooManyRequests:
				if retries >= maxRetries {
					err = fmt.Errorf("maximum (%d) retries exceeded", maxRetries)
					fatal = true
					return
				}
				retryAfter := time.Duration(1<<retries) * time.Second
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					if seconds, err := strconv.Atoi(ra); err == nil {
						retryAfter = time.Duration(seconds) * time.Second
					}
				}
				select {
				case <-time.After(retryAfter):
				case <- ctx.Done():
					err = ctx.Err()
					fatal = true
					return
				}
				err = fmt.Errorf("too many requests: %s", resp.Status)
				retries++

			default:
				if resp.StatusCode >= 500 {
					err = fmt.Errorf("server error: %s", resp.Status)
					retries++
					return
				}
				err = fmt.Errorf("unexpected status: %s", resp.Status)
				fatal = true
			}
		}()

		if fatal {
			return nil, err
		}

		if err == nil && copyErr == nil {
			if totalSize != -1 && written < totalSize {
				continue
			}
			success = true
			return tmp, nil
		}

		if retries >= maxRetries {
			return nil, fmt.Errorf("download failed after %d retries: %w", maxRetries, err)
		}
	}
}


func CreateGetFileRequest(ctx context.Context, cfg *config.ApiConfig, itemID string, written int64) (*http.Request, error) {
	err := checkAccessTokenExpiry(cfg)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/sites/%s/drives/%s/items/%s/content", graphBaseURL, cfg.GraphSiteID, cfg.GraphDriveID, itemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	if written > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", written))
	}

	return req, nil
}


func partialContentPreCopyCheckSum(rangeHeader string, written int64) (int64, int64, int64, error) {
	var start, end, total int64
	
	if rangeHeader == "" {
		return 0, 0, 0, fmt.Errorf("206 response missing Content-Range")
	}
	
	_, err := fmt.Sscanf(rangeHeader, "bytes %d-%d/%d", &start, &end, &total)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse Content-Range: %w", err)
	}

	if start != written {
		return 0, 0, 0, fmt.Errorf("unexpected Content-Range start: %d, expected: %d", start, written)
	}

	if end < start {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range: end: %d < start %d", end, start)
	}

	if total <= 0 {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range total size: %d", total)
	}
	
	return start, end, total, nil
}


