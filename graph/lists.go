package graph

import (
	"net/http"
	"fmt"
	"io"

	"encoding/json"
)

type GraphListResponse[T any] struct {
	Value []T `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}

type RootDir struct {
	ItemID string `json:"id"`
	Name string `json:"name"`
}


const graphBaseURL = "https://graph.microsoft.com/v1.0"











func (cfg *apiConfig) GetRootDirs(ctx context.Context, driveID string) ([]RootDirs, error) {
	url := fmt.Sprintf("%s/drives/%s/root/children", graphBaseURL, driveID)
	
	req, err := NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result, err := cfg.do[GraphListResponse[RootDir]](req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return result.Value, nil
}



func (cfg *apiConfig) GetAllDirs(driveID, itemsID string) (allDirs, error) {
	

}

func (cfg *apiConfig) GetAllUnprocessedDirs(driveID, path string) (error) {

}

func (cfg *apiConfig) Get SingleDir(driveID, path string) (error)  {

}

func (cfg *apiConfig) fetchNextPage[T any](ctx Context.Context, nextLink string) (T, error) {
	continue
}

