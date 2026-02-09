package graph

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/JA50N14/rfp_parser/config"
)


type Parsed struct {
	Parsed bool `json:"Parsed"`
}

func PatchPackageParsed(itemID string, ctx context.Context, cfg *config.ApiConfig) (Parsed, error) {
	err := checkAccessTokenExpiry(cfg)
	if err != nil {
		return Parsed{}, err
	}

	buildReq := func (ctx context.Context) (*http.Request, error) {
		url := fmt.Sprintf("%s/sites/%s/drives/%s/items/%s/listItem/fields", graphBaseURL, cfg.GraphSiteID, cfg.GraphDriveID, itemID)

		body := bytes.NewReader([]byte(`{"Parsed": true}`))

		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, body)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		return req, nil
	}

	result, err := do[Parsed](ctx, cfg, buildReq)
	if err != nil {
		return Parsed{}, err
	}

	return result, nil
}
