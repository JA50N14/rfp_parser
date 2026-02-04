package graph

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/internal/auth"
)


type Item struct {
	ID string `json:"id"`
	Name string `json:"name"`
}

type Package struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Fields struct {
		ListID string `json:"id"`
		ListParsed bool `json:"Parsed"`
		ListContentType string `json:"ContentType"`
	} `json:"fields"`
}

const graphBaseURL = "https://graph.microsoft.com/v1.0"
const refreshTokenWindow = time.Minute * 5


func GetRootDirs(ctx context.Context, cfg *config.ApiConfig) ([]Item, error) {
	err := checkAccessTokenExpiry(cfg)
	if err != nil {
		return nil, err
	}

	buildReq := func (ctx context.Context) (*http.Request, error) {
		url := fmt.Sprintf("%s/drives/%s/root/children", graphBaseURL, cfg.GraphDriveID)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer " + cfg.AccessToken)
		return req, nil
	}

	result, err := listAll[Item](ctx, cfg, buildReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}


func GetItemSubDirs(ctx context.Context, cfg *config.ApiConfig, itemID string) ([]Item, error) {
	err := checkAccessTokenExpiry(cfg)
	if err != nil {
		return nil, err
	}

	buildReq := func (ctx context.Context) (*http.Request, error) {
		url := fmt.Sprintf("%s/drives/%s/items/%s/children", graphBaseURL, cfg.GraphDriveID, itemID)
		
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer " + cfg.AccessToken)
		return req, nil 
	}

	result, err := listAll[Item](ctx, cfg, buildReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}


func GetItemSubDirsWithMetadata(ctx context.Context, cfg *config.ApiConfig, itemID string) ([]Package, error) {
	err := checkAccessTokenExpiry(cfg)
	if err != nil {
		return nil, err
	}

	buildReq := func (ctx context.Context) (*http.Request, error) {
		url := fmt.Sprintf("%s/drives/%s/items/%s/children?expand=listItem", graphBaseURL, cfg.GraphDriveID, itemID)
		
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer " + cfg.AccessToken)
		return req, nil
	}

	result, err := listAll[Package](ctx, cfg, buildReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}






func checkAccessTokenExpiry(cfg *config.ApiConfig) error {
	if time.Until(cfg.AccessTokenExpiresAt) <= refreshTokenWindow {
		tokenResp, err := auth.GetGraphAccessToken(cfg.Client)
		if err != nil {
			return err
		}
		cfg.AccessToken = tokenResp.AccessToken
		cfg.AccessTokenExpiresAt = time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}
	return nil
}





