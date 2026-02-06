package graph

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/JA50N14/rfp_parser/config"
)

//How do I do a checksum/size validation?

func GetFile(ctx context.Context, cfg *config.ApiConfig, itemID string) (*os.File, error) {
	var written int64

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

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := CreateGetFileRequest(ctx, cfg, itemID, written)
		if err != nil {
			return nil, err
		}
		
		resp, err := cfg.Client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("context error: %w", ctx.Err())
			}
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
				if _, err := tmp.Seek(written, io.SeekStart); err != nil {
					copyErr = err
					return
				}
				n, copyErr = io.Copy(tmp, resp.Body)
				written += n

			case http.StatusPartialContent:
				var n int64
				if _, err := tmp.Seek(written, io.SeekStart); err != nil {
					copyErr = err
					return
				}
				n, copyErr = io.Copy(tmp, resp.Body)
				written += n

				//Put into its own function - fileCheckSum()
				if copyErr == nil {
					rangeHeader := resp.Header.Get("Content-Range")
					if rangeHeader == "" {
						return
					}
					var start, end, total int64
					_, err := fmt.Sscanf(rangeHeader, "bytes %d-%d/%d", &start, &end, &total)
					if err != nil {
						return fmt.Errorf("failed to parse Content-Range: %w", err)

					}

				}
				////////////////

			case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
				err = fmt.Errorf("download failed: %s", resp.Status)
				fatal = true
				return

			case http.StatusRequestedRangeNotSatisfiable:
				err = fmt.Errorf("server error: %d", written)
				fatal = true
				return

			default:
				if resp.StatusCode >= 500 {
					err = fmt.Errorf("server error: %s", resp.Status)
					return
				}
				err = fmt.Errorf("unexpected status: %s", resp.Status)
				fatal = true
				return
			}
		}()

		if fatal {
			return nil, err
		}

		if err == nil && copyErr == nil {
			success = true
			return tmp, nil
		}
	}

	return nil, fmt.Errorf("download failed after %d attempts (itemID=%s)", maxRetries, itemID)
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


// func GetFile (ctx context.Context, cfg *config.ApiConfig, itemID string) (*os.File, error)
// 	create written var (int64) starting at 0
// 	create a tmp file
// 	enter for loop -> for attempt := 0; attempt <= maxRetries; attempt++
// 		call CreateGetFileRequest(ctx, cfg, itemID, written) to create the request
// 		make request
// 		check response status
// 			if statusCode == 200 call io.Copy(tmpFile, resp.Body), if io.Copy returns err == nil, then return. If io.Copy return an error, get the number of written bytes and continue onto next iteration. Set "Range" header on next request to "written"
// 				resp.Body.Close()
// 			if statusCode == 206 call io.Copy(tmpFile, resp.Body). Capture number of bytes. If err == nil return. If err != nil, continue onto next iteration to make another request with "Range" header set to "written".
// 				resp.Body.Close()
// 			if ctx.Err() == context.DeadlineExceeded -> return. Let caller decide whether to retry with a new context. 
// 				resp.Body.Close()
// 			if statusCode < 200 || statusCode >= 300 ->return nil, fmt.Errorf("unable to download file. Status Code: %d", resp.StatusCode)
// 				resp.Body.Close()
// 	If for loop exits, return nil, fmt.Errorf("unable to download file. itemID: %s", itemID)

// func CreateGetFileRequest (ctx, context.Context, cfg *config.ApiConfig, itemID string, written int64) (*http.Request, error)
// 	Create the request with a "Range" header set to fmt.Sprintf("bytes=%d-", written"). "written" to be 0 on first request.






