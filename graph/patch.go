package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JA50N14/rfp_parser/config"
)

type ProcessStatus struct {
	ProcessStatus string `json:"ProcessStatus"`
}

func PatchProcessStatus(itemID string, patchValue string, ctx context.Context, cfg *config.ApiConfig) (ProcessStatus, error) {
	err := checkAccessTokenExpiry(cfg)
	if err != nil {
		return ProcessStatus{}, err
	}

	buildReq := func(ctx context.Context) (*http.Request, error) {
		url := fmt.Sprintf("%s/sites/%s/drives/%s/items/%s/listItem/fields", graphBaseURL, cfg.GraphSiteID, cfg.GraphDriveID, itemID)

		payload := map[string]string{
			"ProcessStatus": patchValue,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		body := bytes.NewReader(b)

		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, body)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		return req, nil
	}

	result, err := do[ProcessStatus](ctx, cfg, buildReq)
	if err != nil {
		return ProcessStatus{}, err
	}

	return result, nil
}
