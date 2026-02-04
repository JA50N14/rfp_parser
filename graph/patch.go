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

func PatchPackageParsed(ctx context.Context, cfg *config.ApiConfig, itemID string) (Parsed, error) {
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

		req.Header.Set("Authorization", "bearer " + cfg.AccessToken)

		return req, nil
	}

	result, err := do[Parsed](ctx, cfg, buildReq)
	if err != nil {
		return Parsed{}, err
	}

	return result, nil
}

//Figure out retries with a body. I think you need to recreate the body