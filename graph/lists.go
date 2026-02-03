package graph

import (
	"context"
	"fmt"
	"net/http"

	"github.com/JA50N14/rfp_parser/config"
)


type BusinessUnit struct {
	ItemID string `json:"id"`
	Name string `json:"name"`
}

const graphBaseURL = "https://graph.microsoft.com/v1.0"


func GetRootDirs(ctx context.Context, cfg *config.ApiConfig, driveID string) ([]BusinessUnit, error) {
	url := fmt.Sprintf("%s/drives/%s/root/children", graphBaseURL, cfg.GraphDriveID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	result, err := do[GraphListResponse[RootDir]](req, cfg)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return result.Value, nil
}



func GetItemSubDirs(ctx context.Context, cfg *config.ApiConfig, driveID, itemID string) (allDirs, error) {
	

}


func GetItemSubDirsWithMetadata(ctx context.Context, cfg *config.ApiConfig, driveID, itemID string) (allDirs, error) {

}






// func (cfg *apiConfig) fetchNextPage[T any](ctx Context.Context, nextLink string) (T, error) {
// 	continue
// }

